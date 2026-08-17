<?php

namespace App\Jobs;

use App\Models\Chat\ChatBotProfile;
use App\Models\Chat\ChatBotTemplate;
use App\Models\Chat\ChatMessage;
use App\Models\Chat\ChatRoom;
use App\Services\Chat\ChatRedisPublisher;
use Illuminate\Bus\Queueable;
use Illuminate\Contracts\Queue\ShouldQueue;
use Illuminate\Foundation\Bus\Dispatchable;
use Illuminate\Queue\InteractsWithQueue;
use Illuminate\Queue\SerializesModels;
use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\DB;

class GenerateChatBotMessage implements ShouldQueue
{
    use Dispatchable, InteractsWithQueue, Queueable, SerializesModels;

    public function handle(ChatRedisPublisher $publisher): void
    {
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

                $message = ChatMessage::query()->create([
                    'room_id' => $room->id,
                    'actor_type' => 'bot',
                    'bot_profile_id' => $profile->id,
                    'display_name' => $profile->display_name,
                    'body' => $template->body,
                    'status' => ChatMessage::STATUS_VISIBLE,
                ]);
                $template->forceFill([
                    'last_used_at' => now(),
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
}
