<?php

namespace App\Services\Wheel;

use App\Models\User;
use App\Models\Wheel\WheelAuditLog;
use App\Models\Wheel\WheelCampaign;
use App\Models\Wheel\WheelInvitation;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\DB;
use Illuminate\Validation\ValidationException;

class WheelCampaignService
{
    public function __construct(private readonly WheelEventPublisher $publisher) {}

    /** @param array<int, int|string> $userIds */
    public function inviteUsers(WheelCampaign $campaign, array $userIds, bool $activate = true): int
    {
        $ids = collect($userIds)->map(fn ($id): int => (int) $id)->filter()->unique()->values();
        if ($ids->isEmpty()) {
            throw ValidationException::withMessages(['user_ids' => 'Vui lòng chọn ít nhất một người chơi.']);
        }

        $count = 0;
        DB::transaction(function () use ($campaign, $ids, $activate, &$count): void {
            $campaign = WheelCampaign::query()->with('roundTemplates')->lockForUpdate()->findOrFail($campaign->id);
            $this->validateCampaign($campaign, $activate);
            $users = User::query()->whereIn('id', $ids)->whereIn('role', [2, 4])->get()->keyBy('id');

            foreach ($ids as $userId) {
                if (! $users->has($userId)) {
                    continue;
                }
                $invitation = WheelInvitation::query()->firstOrCreate(
                    ['campaign_id' => $campaign->id, 'user_id' => $userId],
                    ['status' => 'draft'],
                );
                if ($invitation->status !== 'draft') {
                    continue;
                }
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
            if ($invitation->rounds->count() !== 4) {
                $this->snapshotRounds($invitation, $invitation->campaign);
                $invitation->load('rounds');
            }
            if ($invitation->rounds->count() !== 4) {
                throw ValidationException::withMessages(['rounds' => 'Invitation phải có đúng 4 lượt quay.']);
            }
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

    private function validateCampaign(WheelCampaign $campaign, bool $activation): void
    {
        if ($campaign->roundTemplates->count() !== 4) {
            throw ValidationException::withMessages(['rounds' => 'Chiến dịch phải có đúng 4 lượt mẫu.']);
        }
        if ($activation && $campaign->status !== 'active') {
            throw ValidationException::withMessages(['status' => 'Chiến dịch phải ở trạng thái đang mở.']);
        }
        if ($activation && $campaign->closes_at && $campaign->closes_at->isPast()) {
            throw ValidationException::withMessages(['closes_at' => 'Chiến dịch đã hết thời gian nhận người chơi.']);
        }
    }

    private function snapshotRounds(WheelInvitation $invitation, WheelCampaign $campaign): void
    {
        foreach ($campaign->roundTemplates as $template) {
            $invitation->rounds()->updateOrCreate(['round_no' => $template->round_no], [
                'segment_key' => $template->segment_key,
                'result_label' => $template->result_label,
                'prize_amount' => $template->prize_amount,
                'status' => 'pending',
            ]);
        }
    }

    private function activateLocked(WheelInvitation $invitation, WheelCampaign $campaign): void
    {
        $invitation->forceFill([
            'status' => 'pending',
            'activated_at' => now(),
            'expires_at' => $campaign->closes_at,
            'activated_by' => Auth::id(),
        ])->save();
        $payload = ['invitation_id' => $invitation->public_id, 'campaign_name' => $campaign->name, 'expires_at' => $invitation->expires_at?->toISOString()];
        $this->audit($invitation, 'invitation.activated', ['status' => 'draft'], $payload);
        $this->publisher->queueForUser((int) $invitation->user_id, 'wheel.invitation.activated', $payload);
        DB::afterCommit(fn () => $this->publisher->publishPending());
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
