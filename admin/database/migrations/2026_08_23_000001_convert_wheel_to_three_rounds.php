<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Support\Facades\DB;

return new class extends Migration
{
    public function up(): void
    {
        DB::transaction(function (): void {
            DB::table('wheel_campaign_round_templates')
                ->where('round_no', 2)
                ->update([
                    'segment_key' => 'reward_39m',
                    'result_label' => '39 triệu',
                    'prize_amount' => 39000000,
                    'updated_at' => now('UTC'),
                ]);
            DB::table('wheel_campaign_round_templates')->where('round_no', '>', 3)->delete();

            DB::table('wheel_invitation_rounds')
                ->where('round_no', 2)
                ->where('status', 'pending')
                ->whereIn('invitation_id', DB::table('wheel_invitations')->select('id')->whereIn('status', ['draft', 'pending', 'started']))
                ->update([
                    'segment_key' => 'reward_39m',
                    'result_label' => '39 triệu',
                    'prize_amount' => 39000000,
                    'updated_at' => now('UTC'),
                ]);

            DB::table('wheel_invitation_rounds')
                ->where('round_no', '>', 3)
                ->where('status', '<>', 'spun')
                ->whereIn('invitation_id', DB::table('wheel_invitations')->select('id')->whereIn('status', ['draft', 'pending', 'started']))
                ->delete();

            $completedSessionIDs = DB::table('wheel_sessions')
                ->where('status', 'active')
                ->where('current_round', '>', 3)
                ->pluck('id');
            if ($completedSessionIDs->isEmpty()) {
                return;
            }

            $now = now('UTC');
            $invitationIDs = DB::table('wheel_sessions')->whereIn('id', $completedSessionIDs)->pluck('invitation_id');
            DB::table('wheel_sessions')->whereIn('id', $completedSessionIDs)->update([
                'status' => 'completed',
                'current_round' => 3,
                'completed_at' => $now,
                'version' => DB::raw('version + 1'),
                'updated_at' => $now,
            ]);
            DB::table('wheel_invitations')->whereIn('id', $invitationIDs)->update(['status' => 'completed', 'updated_at' => $now]);
            DB::table('chat_rooms')->whereIn('wheel_session_id', $completedSessionIDs)->update(['enabled' => false, 'updated_at' => $now]);
        });
    }

    public function down(): void
    {
        // Historical results cannot be reconstructed safely.
    }
};
