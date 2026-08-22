<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Support\Facades\DB;

return new class extends Migration
{
    public function up(): void
    {
        DB::statement('CREATE INDEX IF NOT EXISTS chat_messages_actor_type_id_index ON chat_messages (actor_type, id DESC)');

        DB::statement(<<<'SQL'
            UPDATE chat_messages AS cm
            SET display_name = 'ID game #' || cm.user_id::text,
                updated_at = timezone('UTC', now())
            FROM chat_rooms AS cr
            WHERE cm.room_id = cr.id
              AND cm.actor_type = 'user'
              AND cm.user_id IS NOT NULL
              AND (cr.wheel_invitation_id IS NOT NULL OR cr.wheel_session_id IS NOT NULL)
        SQL);
    }

    public function down(): void
    {
        DB::statement('DROP INDEX IF EXISTS chat_messages_actor_type_id_index');
    }
};
