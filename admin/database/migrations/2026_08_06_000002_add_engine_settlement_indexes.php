<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        if (DB::getDriverName() !== 'pgsql') {
            return;
        }

        if (Schema::hasTable('game_periods')) {
            DB::statement(
                'CREATE INDEX IF NOT EXISTS "idx_game_periods_status_draw_at_id" '
                .'ON "game_periods" ("status", "draw_at", "id")'
            );
        }

        if (Schema::hasTable('bet_tickets')) {
            DB::statement(
                'CREATE INDEX IF NOT EXISTS "idx_bet_tickets_wallet_status" '
                .'ON "bet_tickets" ("wallet_id", "status")'
            );
        }
    }

    public function down(): void
    {
        if (DB::getDriverName() !== 'pgsql') {
            return;
        }

        DB::statement('DROP INDEX IF EXISTS "idx_bet_tickets_wallet_status"');
        DB::statement('DROP INDEX IF EXISTS "idx_game_periods_status_draw_at_id"');
    }
};
