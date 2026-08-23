<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('chat_rooms', function (Blueprint $table): void {
            $table->timestamp('bot_active_until')->nullable()->after('next_bot_at');
            $table->index(['enabled', 'bot_active_until']);
        });

        if (DB::getDriverName() === 'pgsql') {
            DB::statement(<<<'SQL'
                UPDATE chat_rooms AS cr
                SET bot_active_until = ws.ends_at,
                    next_bot_at = timezone('UTC', now()),
                    updated_at = timezone('UTC', now())
                FROM wheel_sessions AS ws
                JOIN wheel_invitations AS wi ON wi.id = ws.invitation_id
                WHERE cr.wheel_session_id = ws.id
                  AND cr.enabled = true
                  AND ws.status = 'active'
                  AND ws.ends_at > timezone('UTC', now())
                  AND wi.bot_chat_enabled = true
            SQL);
        }
    }

    public function down(): void
    {
        Schema::table('chat_rooms', function (Blueprint $table): void {
            $table->dropIndex(['enabled', 'bot_active_until']);
            $table->dropColumn('bot_active_until');
        });
    }
};
