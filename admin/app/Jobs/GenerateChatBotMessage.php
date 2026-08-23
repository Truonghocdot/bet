<?php

namespace App\Jobs;

use App\Models\Chat\ChatBotProfile;
use App\Models\Chat\ChatBotTemplate;
use App\Models\Chat\ChatMessage;
use App\Models\Chat\ChatRoom;
use App\Services\Chat\ChatRedisPublisher;
use Carbon\CarbonImmutable;
use Illuminate\Bus\Queueable;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Bus\Dispatchable;
use Illuminate\Queue\InteractsWithQueue;
use Illuminate\Queue\SerializesModels;
use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Log;

class GenerateChatBotMessage implements ShouldQueue
{
    use Dispatchable, InteractsWithQueue, Queueable, SerializesModels;

    public const INITIAL_BURST_COUNT = 4;

    public const MIN_INTERVAL_SECONDS = 1;

    public const MAX_INTERVAL_SECONDS = 3;

    public function __construct(public readonly ?int $wheelRoomId = null) {}

    public function handle(ChatRedisPublisher $publisher): void
    {
        if (config('wheel.enabled')) {
            if ($this->wheelRoomId !== null) {
                $this->generateForWheelRoom($publisher, $this->wheelRoomId);

                return;
            }

            if ($this->generateForWheelSession($publisher)) {
                return;
            }
        }

        if (! filter_var(env('CHAT_GLOBAL_ENABLED', false), FILTER_VALIDATE_BOOL)) {
            return;
        }

        $lock = Cache::lock('chat:bot:generate:global', 30);
        if (! $lock->get()) {
            return;
        }

        try {
            $room = ChatRoom::query()->where('code', env('CHAT_ROOM_CODE', 'global'))->where('enabled', true)->first();
            if (! $room) {
                return;
            }

            $message = DB::transaction(function () use ($room): ?ChatMessage {
                $profile = ChatBotProfile::query()->where('active', true)->inRandomOrder()->first();
                if (! $profile) {
                    return null;
                }

                $lastBotBody = ChatMessage::query()
                    ->where('room_id', $room->id)
                    ->where('actor_type', 'bot')
                    ->latest('id')
                    ->value('body');

                $template = ChatBotTemplate::query()
                    ->where('active', true)
                    ->where(function ($query) use ($profile): void {
                        $query->whereNull('bot_profile_id')->orWhere('bot_profile_id', $profile->id);
                    })
                    ->when($lastBotBody, fn ($query) => $query->where('body', '<>', $lastBotBody))
                    ->orderByRaw('last_used_at is not null')
                    ->orderBy('last_used_at')
                    ->orderBy('id')
                    ->lockForUpdate()
                    ->first();
                if (! $template) {
                    return null;
                }

                $timestamp = now('UTC');
                $message = ChatMessage::query()->forceCreate([
                    'room_id' => $room->id,
                    'actor_type' => 'bot',
                    'bot_profile_id' => $profile->id,
                    'display_name' => $profile->display_name,
                    'body' => $template->body,
                    'status' => ChatMessage::STATUS_VISIBLE,
                    'created_at' => $timestamp,
                    'updated_at' => $timestamp,
                ]);
                $template->forceFill([
                    'last_used_at' => now('UTC'),
                    'usage_count' => ((int) $template->usage_count) + 1,
                ])->save();

                return $message;
            });

            if ($message) {
                $publisher->publish($room->code, 'chat.message.created', $message);
            }
        } finally {
            optional($lock)->release();
        }
    }

    private function generateForWheelSession(ChatRedisPublisher $publisher, ?int $onlyRoomId = null): bool
    {
        $now = now('UTC');
        $room = ChatRoom::query()
            ->where('enabled', true)
            ->when($onlyRoomId !== null, fn ($query) => $query->whereKey($onlyRoomId))
            ->where(fn ($query) => $query
                ->whereNull('next_bot_at')
                ->orWhere('next_bot_at', '<=', $now)
                ->orWhere('bot_message_count', '<', self::INITIAL_BURST_COUNT))
            ->where(function ($query) use ($now): void {
                $query
                    ->where(function ($query) use ($now): void {
                        $query->whereNotNull('wheel_session_id')
                            ->whereHas('wheelSession', fn ($sessionQuery) => $sessionQuery
                                ->where('status', 'active')
                                ->where('ends_at', '>', $now)
                                ->whereHas('invitation', fn ($invitationQuery) => $invitationQuery->where('bot_chat_enabled', true)));
                    })
                    ->orWhere(function ($query): void {
                        $query->whereNull('wheel_session_id')
                            ->whereNotNull('wheel_invitation_id')
                            ->where(fn ($pendingQuery) => $pendingQuery
                                ->where('bot_message_count', '<', 30)
                                ->orWhere('bot_active_until', '>', now('UTC')))
                            ->whereHas('wheelInvitation', fn ($invitationQuery) => $invitationQuery
                                ->where('status', 'pending')
                                ->where('bot_chat_enabled', true));
                    });
            })
            ->with('wheelSession')
            ->orderBy('next_bot_at')
            ->first();
        if (! $room) {
            return false;
        }

        $lock = Cache::lock('chat:bot:wheel-room:'.$room->id, 20);
        if (! $lock->get()) {
            return true;
        }

        try {
            $messages = DB::transaction(function () use ($room): array {
                $lockedRoom = ChatRoom::query()->with(['wheelSession', 'wheelInvitation'])->lockForUpdate()->find($room->id);
                $lockedNextAt = $lockedRoom?->getRawOriginal('next_bot_at');
                if ($lockedNextAt
                    && CarbonImmutable::parse($lockedNextAt, 'UTC')->isFuture()
                    && ((int) $lockedRoom?->bot_message_count >= self::INITIAL_BURST_COUNT)) {
                    return [];
                }
                $rawEndsAt = $lockedRoom?->wheelSession?->getRawOriginal('ends_at');
                $endsAtUtc = $rawEndsAt ? CarbonImmutable::parse($rawEndsAt, 'UTC') : null;
                $sessionActive = $lockedRoom?->wheelSession
                    && $lockedRoom->wheelSession->status === 'active'
                    && $endsAtUtc?->isFuture()
                    && $lockedRoom->wheelInvitation?->bot_chat_enabled;
                $rawBotActiveUntil = $lockedRoom?->getRawOriginal('bot_active_until');
                $pendingWindowActive = $rawBotActiveUntil
                    && CarbonImmutable::parse($rawBotActiveUntil, 'UTC')->isFuture();
                $invitationPending = $lockedRoom?->wheel_session_id === null
                    && $lockedRoom->wheelInvitation?->status === 'pending'
                    && $lockedRoom->wheelInvitation?->bot_chat_enabled
                    && ($pendingWindowActive || (int) $lockedRoom->bot_message_count < 30);
                if (! $lockedRoom || ! $lockedRoom->enabled || (! $sessionActive && ! $invitationPending)) {
                    return [];
                }

                $remainingBurst = max(0, self::INITIAL_BURST_COUNT - (int) $lockedRoom->bot_message_count);
                $generateCount = max(1, $remainingBurst);
                if ($invitationPending && ! $pendingWindowActive) {
                    $generateCount = min($generateCount, max(0, 30 - (int) $lockedRoom->bot_message_count));
                }

                $messages = [];
                $lastProfileId = null;
                for ($index = 0; $index < $generateCount; $index++) {
                    $profile = ChatBotProfile::query()
                        ->where('active', true)
                        ->when($lastProfileId, fn ($query) => $query->where('id', '<>', $lastProfileId))
                        ->inRandomOrder()
                        ->first();
                    if (! $profile) {
                        Log::warning('Wheel bot skipped: no active bot profile.', ['room_id' => $lockedRoom->id]);
                        break;
                    }
                    $lastProfileId = (int) $profile->id;

                    $lastBotBody = ChatMessage::query()->where('room_id', $lockedRoom->id)->where('actor_type', 'bot')->latest('id')->value('body');
                    $template = ChatBotTemplate::query()
                        ->where('active', true)
                        ->where(function ($query) use ($profile): void {
                            $query->whereNull('bot_profile_id')->orWhere('bot_profile_id', $profile->id);
                        })
                        ->when($lastBotBody, fn ($query) => $query->where('body', '<>', $lastBotBody))
                        ->orderByRaw('last_used_at is not null')
                        ->orderBy('last_used_at')
                        ->orderBy('id')
                        ->lockForUpdate()
                        ->first();
                    if (! $template) {
                        $template = ChatBotTemplate::query()
                            ->where('active', true)
                            ->when($lastBotBody, fn ($query) => $query->where('body', '<>', $lastBotBody))
                            ->orderByRaw('last_used_at is not null')
                            ->orderBy('last_used_at')
                            ->orderBy('id')
                            ->lockForUpdate()
                            ->first();
                    }
                    if (! $template) {
                        Log::warning('Wheel bot skipped: no active bot template.', ['room_id' => $lockedRoom->id]);
                        break;
                    }

                    $timestamp = now('UTC');
                    $message = ChatMessage::query()->forceCreate([
                        'room_id' => $lockedRoom->id,
                        'actor_type' => 'bot',
                        'bot_profile_id' => $profile->id,
                        'display_name' => $profile->display_name,
                        'body' => $template->body,
                        'status' => ChatMessage::STATUS_VISIBLE,
                        'created_at' => $timestamp,
                        'updated_at' => $timestamp,
                    ]);
                    $template->forceFill(['last_used_at' => $timestamp, 'usage_count' => ((int) $template->usage_count) + 1])->save();
                    $messages[] = $message;
                }

                $lockedRoom->forceFill([
                    'next_bot_at' => now('UTC')->addSeconds(random_int(self::MIN_INTERVAL_SECONDS, self::MAX_INTERVAL_SECONDS)),
                    'bot_message_count' => ((int) $lockedRoom->bot_message_count) + count($messages),
                ])->save();

                return $messages;
            });

            $room->refresh();
            foreach ($messages as $message) {
                if ($room->wheel_session_id) {
                    $publisher->publishWheelSession((int) $room->wheel_session_id, 'chat.message.created', $message);
                } elseif ($room->wheel_invitation_id) {
                    $publisher->publishWheelInvitation((int) $room->wheel_invitation_id, 'chat.message.created', $message);
                }
            }

            if ($room->wheel_invitation_id) {
                $this->queueNextWheelMessage((int) $room->id);
            }

            return true;
        } finally {
            optional($lock)->release();
        }
    }

    private function generateForWheelRoom(ChatRedisPublisher $publisher, int $roomId): void
    {
        $room = ChatRoom::query()->whereKey($roomId)->where('enabled', true)->first();
        if (! $room) {
            return;
        }

        $rawNextAt = $room->getRawOriginal('next_bot_at');
        if ($rawNextAt
            && CarbonImmutable::parse($rawNextAt, 'UTC')->isFuture()
            && ((int) $room->bot_message_count >= self::INITIAL_BURST_COUNT)) {
            $this->queueNextWheelMessage($roomId);

            return;
        }

        $this->generateForWheelSession($publisher, $roomId);
    }

    private function queueNextWheelMessage(int $roomId): void
    {
        $room = ChatRoom::query()->with(['wheelSession', 'wheelInvitation'])->find($roomId);
        if (! $room || ! $room->enabled || ! $room->wheelInvitation?->bot_chat_enabled) {
            return;
        }

        $sessionActive = false;
        if ($room->wheelSession) {
            $rawEndsAt = $room->wheelSession->getRawOriginal('ends_at');
            $sessionActive = $room->wheelSession->status === 'active'
                && $rawEndsAt
                && CarbonImmutable::parse($rawEndsAt, 'UTC')->isFuture();
        }
        $invitationPending = $room->wheel_session_id === null
            && $room->wheelInvitation->status === 'pending'
            && (
                (int) $room->bot_message_count < 30
                || ($room->getRawOriginal('bot_active_until')
                    && CarbonImmutable::parse($room->getRawOriginal('bot_active_until'), 'UTC')->isFuture())
            );
        if (! $sessionActive && ! $invitationPending) {
            return;
        }

        $rawNextAt = $room->getRawOriginal('next_bot_at');
        $nextAt = $rawNextAt ? CarbonImmutable::parse($rawNextAt, 'UTC') : now('UTC');
        $delay = max(1, $nextAt->getTimestamp() - now('UTC')->getTimestamp());
        $scheduleKey = 'chat:bot:wheel:scheduled:'.$roomId.':'.$nextAt->getTimestamp();
        if (! Cache::add($scheduleKey, 1, $delay + 30)) {
            return;
        }
        self::dispatch($roomId)->delay($delay);
    }
}
