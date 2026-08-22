<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Support\Facades\DB;

return new class extends Migration
{
    public function up(): void
    {
        DB::statement('ALTER TABLE wheel_invitations ALTER COLUMN bot_chat_enabled SET DEFAULT true');

        // Every active invitation gets a room so chat can open before the
        // wheel session. The bot flag controls generation, not room access.
        DB::statement(<<<'SQL'
            INSERT INTO chat_rooms
                (wheel_invitation_id, code, name, enabled, next_bot_at, bot_message_count, created_at, updated_at)
            SELECT wi.id,
                   'wheel-invitation-' || wi.id,
                   'Phòng sự kiện ' || wc.name,
                   true,
                   CASE WHEN wi.bot_chat_enabled
                        THEN timezone('UTC', now()) + ((8 + floor(random() * 7)) * interval '1 second')
                        ELSE null
                   END,
                   0,
                   timezone('UTC', now()),
                   timezone('UTC', now())
            FROM wheel_invitations AS wi
            JOIN wheel_campaigns AS wc ON wc.id = wi.campaign_id
            WHERE wi.status = 'pending'
              AND NOT EXISTS (
                  SELECT 1 FROM chat_rooms AS existing
                  WHERE existing.wheel_invitation_id = wi.id
              )
        SQL);

        // Old sessions created while bot chat was disabled have a room linked
        // only to the session. Attach it to the invitation for pre-session and
        // invitation-scoped chat lookups.
        DB::statement(<<<'SQL'
            UPDATE chat_rooms AS cr
            SET wheel_invitation_id = ws.invitation_id,
                updated_at = timezone('UTC', now())
            FROM wheel_sessions AS ws
            WHERE cr.wheel_session_id = ws.id
              AND cr.wheel_invitation_id IS NULL
              AND NOT EXISTS (
                  SELECT 1 FROM chat_rooms AS existing
                  WHERE existing.wheel_invitation_id = ws.invitation_id
              )
        SQL);
    }

    public function down(): void
    {
        DB::statement('ALTER TABLE wheel_invitations ALTER COLUMN bot_chat_enabled SET DEFAULT false');
    }
};
