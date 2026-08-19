<?php

namespace App\Console\Commands;

use App\Services\Wheel\WheelEventPublisher;
use Illuminate\Console\Command;
use Illuminate\Support\Facades\DB;

class MaintainWheelEventsCommand extends Command
{
    protected $signature = 'wheel:maintain';

    protected $description = 'Đóng campaign và hết hạn invitation/session vòng quay';

    public function handle(WheelEventPublisher $publisher): int
    {
        if (! config('wheel.enabled')) {
            return self::SUCCESS;
        }

        DB::transaction(function () use ($publisher): void {
            DB::table('wheel_campaigns')->where('status', 'draft')->whereNotNull('opens_at')->where('opens_at', '<=', now())->update(['status' => 'active', 'updated_at' => now()]);
            DB::table('wheel_campaigns')->where('status', 'active')->whereNotNull('closes_at')->where('closes_at', '<=', now())->update(['status' => 'closed', 'updated_at' => now()]);

            foreach (DB::table('wheel_invitations')->where('status', 'pending')->whereNotNull('expires_at')->where('expires_at', '<=', now())->lockForUpdate()->get() as $invitation) {
                DB::table('wheel_invitations')->where('id', $invitation->id)->update(['status' => 'expired', 'updated_at' => now()]);
                DB::table('wheel_audit_logs')->insert(['campaign_id' => $invitation->campaign_id, 'invitation_id' => $invitation->id, 'action' => 'invitation.expired', 'new_values' => json_encode(['status' => 'expired'], JSON_THROW_ON_ERROR), 'created_at' => now()]);
                $publisher->queueForUser((int) $invitation->user_id, 'wheel.invitation.expired', ['invitation_id' => $invitation->public_id]);
            }

            foreach (DB::table('wheel_sessions')->where('status', 'active')->where('ends_at', '<=', now())->lockForUpdate()->get() as $session) {
                DB::table('wheel_sessions')->where('id', $session->id)->update(['status' => 'expired', 'expired_at' => now(), 'updated_at' => now()]);
                DB::table('wheel_invitations')->where('id', $session->invitation_id)->update(['status' => 'expired', 'updated_at' => now()]);
                DB::table('wheel_invitation_rounds')->where('invitation_id', $session->invitation_id)->where('status', 'pending')->update(['status' => 'expired', 'updated_at' => now()]);
                DB::table('chat_rooms')->where('wheel_session_id', $session->id)->update(['enabled' => false, 'updated_at' => now()]);
                DB::table('wheel_audit_logs')->insert(['invitation_id' => $session->invitation_id, 'session_id' => $session->id, 'action' => 'session.expired', 'new_values' => json_encode(['status' => 'expired'], JSON_THROW_ON_ERROR), 'created_at' => now()]);
                $publisher->queueForSession((int) $session->id, 'wheel.session.expired', ['session_id' => $session->public_id]);
            }
        });

        $publisher->publishPending();

        return self::SUCCESS;
    }
}
