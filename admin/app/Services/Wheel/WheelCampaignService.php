<?php

namespace App\Services\Wheel;

use App\Jobs\GenerateChatBotMessage;
use App\Models\Chat\ChatRoom;
use App\Models\User;
use App\Models\Wheel\WheelAuditLog;
use App\Models\Wheel\WheelCampaign;
use App\Models\Wheel\WheelInvitation;
use App\Support\Decimal;
use Illuminate\Support\Collection;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\DB;
use Illuminate\Validation\ValidationException;

class WheelCampaignService
{
    public const TOTAL_ROUNDS = 3;

    public const SECOND_ROUND_PRIZE = '39000000.00000000';

    public function __construct(private readonly WheelEventPublisher $publisher) {}

    /** @param array<int, int|string> $userIds */
    public function inviteUsers(WheelCampaign $campaign, array $userIds, bool $activate = true, bool $botChatEnabled = true): int
    {
        $ids = collect($userIds)->map(fn ($id): int => (int) $id)->filter()->unique()->values();
        if ($ids->isEmpty()) {
            throw ValidationException::withMessages(['user_ids' => 'Vui lòng chọn ít nhất một người chơi.']);
        }

        $count = 0;
        DB::transaction(function () use ($campaign, $ids, $activate, $botChatEnabled, &$count): void {
            $campaign = WheelCampaign::query()->with('roundTemplates')->lockForUpdate()->findOrFail($campaign->id);
            $this->validateCampaign($campaign, $activate);
            $users = User::query()->whereIn('id', $ids)->whereIn('role', [2, 4])->get()->keyBy('id');

            foreach ($ids as $userId) {
                if (! $users->has($userId)) {
                    continue;
                }
                $invitation = WheelInvitation::query()->create([
                    'campaign_id' => $campaign->id,
                    'user_id' => $userId,
                    'status' => 'draft',
                    'bot_chat_enabled' => $botChatEnabled,
                ]);
                $this->snapshotRounds($invitation, $campaign);
                if ($activate) {
                    $this->activateLocked($invitation, $campaign);
                }
                $count++;
            }
        });

        return $count;
    }

    public function snapshotInvitation(WheelInvitation $invitation): void
    {
        DB::transaction(function () use ($invitation): void {
            $invitation = WheelInvitation::query()->with('campaign.roundTemplates')->lockForUpdate()->findOrFail($invitation->id);
            if ($invitation->status === 'draft') {
                $this->snapshotRounds($invitation, $invitation->campaign);
            }
        });
    }

    public function activate(WheelInvitation $invitation): void
    {
        DB::transaction(function () use ($invitation): void {
            $invitation = WheelInvitation::query()->with(['campaign.roundTemplates', 'rounds'])->lockForUpdate()->findOrFail($invitation->id);
            if ($invitation->status !== 'draft') {
                return;
            }
            $this->validateCampaign($invitation->campaign, true);
            if ($invitation->rounds->count() !== self::TOTAL_ROUNDS) {
                $this->snapshotRounds($invitation, $invitation->campaign);
                $invitation->load('rounds');
            }
            $this->validateRoundSet($invitation->rounds, 'rounds', 'Lời mời');
            $this->activateLocked($invitation, $invitation->campaign);
        });
    }

    public function revoke(WheelInvitation $invitation): void
    {
        DB::transaction(function () use ($invitation): void {
            $invitation = WheelInvitation::query()->lockForUpdate()->findOrFail($invitation->id);
            if (! in_array($invitation->status, ['draft', 'pending'], true)) {
                throw ValidationException::withMessages(['status' => 'Chỉ có thể thu hồi lời mời chưa bắt đầu.']);
            }
            $old = $invitation->status;
            $invitation->forceFill(['status' => 'revoked', 'revoked_at' => now(), 'revoked_by' => Auth::id()])->save();
            $this->audit($invitation, 'invitation.revoked', ['status' => $old], ['status' => 'revoked']);
            $this->publisher->queueForUser((int) $invitation->user_id, 'wheel.invitation.revoked', ['invitation_id' => $invitation->public_id]);
            DB::afterCommit(fn () => $this->publisher->publishPending());
        });
    }

    public function setBotChatEnabled(WheelInvitation $invitation, bool $enabled): void
    {
        $roomId = null;

        DB::transaction(function () use ($invitation, $enabled, &$roomId): void {
            $invitation = WheelInvitation::query()
                ->with(['campaign', 'session'])
                ->lockForUpdate()
                ->findOrFail($invitation->id);

            if (! in_array($invitation->status, ['pending', 'started'], true)) {
                throw ValidationException::withMessages([
                    'status' => 'Chỉ có thể thay đổi bot chat khi lời mời đang chờ hoặc phiên đang chạy.',
                ]);
            }

            $old = (bool) $invitation->bot_chat_enabled;
            $invitation->forceFill(['bot_chat_enabled' => $enabled])->save();
            $room = $this->ensureInvitationChatRoom($invitation, $invitation->campaign, $enabled);
            $roomId = (int) $room->id;
            $this->audit(
                $invitation,
                $enabled ? 'invitation.bot_chat_enabled' : 'invitation.bot_chat_disabled',
                ['bot_chat_enabled' => $old],
                ['bot_chat_enabled' => $enabled],
            );
        });

        if ($enabled && $roomId) {
            GenerateChatBotMessage::dispatch($roomId);
        }
    }

    private function validateCampaign(WheelCampaign $campaign, bool $activation): void
    {
        $this->validateRoundSet($campaign->roundTemplates, 'rounds', 'Chiến dịch');
        if ($activation && $campaign->status !== 'active') {
            throw ValidationException::withMessages(['status' => 'Chiến dịch phải ở trạng thái đang mở.']);
        }
    }

    private function snapshotRounds(WheelInvitation $invitation, WheelCampaign $campaign): void
    {
        $invitation->rounds()->where('round_no', '>', self::TOTAL_ROUNDS)->delete();
        foreach ($campaign->roundTemplates as $template) {
            $invitation->rounds()->updateOrCreate(['round_no' => $template->round_no], [
                'segment_key' => $template->segment_key,
                'result_label' => $template->result_label,
                'prize_amount' => $template->prize_amount,
                'status' => 'pending',
            ]);
        }
    }

    private function validateRoundSet(Collection $rounds, string $field, string $subject): void
    {
        $numbers = $rounds->pluck('round_no')->map(fn ($value): int => (int) $value)->sort()->values()->all();
        if ($numbers !== range(1, self::TOTAL_ROUNDS)) {
            throw ValidationException::withMessages([$field => "{$subject} phải có đúng 3 lượt quay theo thứ tự 1, 2, 3."]);
        }

        $secondRound = $rounds->firstWhere('round_no', 2);
        $amount = Decimal::normalize((string) $secondRound?->prize_amount);
        if ($amount !== self::SECOND_ROUND_PRIZE || $secondRound?->segment_key !== 'reward_39m') {
            throw ValidationException::withMessages([$field => 'Lượt 2 bắt buộc là giải 39.000.000 VND (mã ô reward_39m).']);
        }
    }

    private function activateLocked(WheelInvitation $invitation, WheelCampaign $campaign): void
    {
        $invitation->forceFill([
            'status' => 'pending',
            'activated_at' => now(),
            'expires_at' => null,
            'activated_by' => Auth::id(),
        ])->save();
        $payload = ['invitation_id' => $invitation->public_id, 'campaign_name' => $campaign->name, 'expires_at' => $invitation->expires_at?->toISOString()];
        $this->audit($invitation, 'invitation.activated', ['status' => 'draft'], $payload);
        $this->publisher->queueForUser((int) $invitation->user_id, 'wheel.invitation.activated', $payload);
        $room = $this->ensureInvitationChatRoom($invitation, $campaign, (bool) $invitation->bot_chat_enabled);
        if ($invitation->bot_chat_enabled) {
            DB::afterCommit(fn () => GenerateChatBotMessage::dispatch((int) $room->id));
        }
        DB::afterCommit(fn () => $this->publisher->publishPending());
    }

    private function ensureInvitationChatRoom(WheelInvitation $invitation, WheelCampaign $campaign, bool $botEnabled): ChatRoom
    {
        $room = ChatRoom::query()->where('wheel_invitation_id', $invitation->id)->first();
        if (! $room && $invitation->session?->id) {
            $room = ChatRoom::query()->where('wheel_session_id', $invitation->session->id)->first();
        }
        $room ??= new ChatRoom;

        $room->forceFill([
            'wheel_invitation_id' => $invitation->id,
            'wheel_session_id' => $invitation->session?->id,
            'code' => $room->exists ? $room->code : 'wheel-invitation-'.$invitation->id,
            'name' => 'Phòng sự kiện '.$campaign->name,
            'enabled' => true,
            // PostgreSQL stores event timestamps as UTC wall-clock values.
            'next_bot_at' => $botEnabled ? now('UTC') : null,
            'bot_active_until' => $botEnabled && $invitation->session
                ? $invitation->session->getRawOriginal('ends_at')
                : null,
        ])->save();

        return $room;
    }

    private function audit(WheelInvitation $invitation, string $action, ?array $old, ?array $new): void
    {
        WheelAuditLog::query()->create([
            'campaign_id' => $invitation->campaign_id,
            'invitation_id' => $invitation->id,
            'actor_user_id' => Auth::id(),
            'action' => $action,
            'old_values' => $old,
            'new_values' => $new,
            'ip_address' => request()?->ip(),
            'created_at' => now(),
        ]);
    }
}
