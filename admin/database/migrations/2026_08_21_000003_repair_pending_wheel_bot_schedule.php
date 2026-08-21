<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Support\Facades\DB;

return new class extends Migration
{
    public function up(): void
    {
        // The first rollout wrote next_bot_at through Laravel's local timezone
        // while PostgreSQL/Gin compare timestamp-without-timezone values as
        // UTC. Reset only pre-session event rooms; global chat and live rooms
        // are intentionally untouched.
        DB::statement(<<<'SQL'
            INSERT INTO chat_rooms
                (wheel_invitation_id, code, name, enabled, next_bot_at, bot_message_count, created_at, updated_at)
            SELECT wi.id,
                   'wheel-invitation-' || wi.id,
                   'Phòng sự kiện ' || wc.name,
                   true,
                   timezone('UTC', now()) + ((8 + floor(random() * 7)) * interval '1 second'),
                   0,
                   timezone('UTC', now()),
                   timezone('UTC', now())
            FROM wheel_invitations AS wi
            JOIN wheel_campaigns AS wc ON wc.id = wi.campaign_id
            WHERE wi.status = 'pending'
              AND wi.bot_chat_enabled = true
              AND NOT EXISTS (
                  SELECT 1 FROM chat_rooms AS existing
                  WHERE existing.wheel_invitation_id = wi.id
              )
        SQL);

        DB::statement(<<<'SQL'
            UPDATE chat_rooms AS cr
            SET next_bot_at = timezone('UTC', now()) + ((8 + floor(random() * 7)) * interval '1 second'),
                updated_at = timezone('UTC', now())
            FROM wheel_invitations AS wi
            WHERE cr.wheel_invitation_id = wi.id
              AND cr.wheel_session_id IS NULL
              AND cr.enabled = true
              AND wi.status = 'pending'
              AND wi.bot_chat_enabled = true
        SQL);
    }

    public function down(): void
    {
        // Repair is intentionally not reversed: restoring the old timezone
        // offset would make pending bot rooms unavailable again.
    }
};
