<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Support\Facades\DB;

return new class extends Migration
{
    public function up(): void
    {
        // Wheel chat rows were historically written as Vietnam wall-clock
        // timestamps, then Gin labelled them as UTC. Vietnam has no DST, so
        // normalize existing event messages once before all writers use UTC.
        DB::statement(<<<'SQL'
            UPDATE chat_messages AS cm
            SET created_at = cm.created_at - interval '7 hours',
                updated_at = cm.updated_at - interval '7 hours'
            FROM chat_rooms AS cr
            WHERE cm.room_id = cr.id
              AND (cr.wheel_invitation_id IS NOT NULL OR cr.wheel_session_id IS NOT NULL)
        SQL);

        // Make every enabled event room immediately eligible. The worker
        // seeds four messages atomically, then schedules the 3-6 second flow.
        DB::statement(<<<'SQL'
            UPDATE chat_rooms AS cr
            SET next_bot_at = timezone('UTC', now()),
                updated_at = timezone('UTC', now())
            FROM wheel_invitations AS wi
            WHERE cr.wheel_invitation_id = wi.id
              AND cr.enabled = true
              AND wi.bot_chat_enabled = true
              AND wi.status IN ('pending', 'started')
        SQL);
    }

    public function down(): void
    {
        // Timestamp normalization is intentionally irreversible.
    }
};
