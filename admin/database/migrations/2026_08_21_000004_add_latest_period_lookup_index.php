<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Support\Facades\DB;

return new class extends Migration
{
    // Building this index concurrently keeps period generation and settlement
    // writes available while PostgreSQL scans the existing table.
    public $withinTransaction = false;

    public function up(): void
    {
        if (DB::getDriverName() !== 'pgsql') {
            return;
        }

        // EnsureRoomPeriods looks up the newest period for one room and locks
        // that row. The existing (room_code, status, draw_at) index cannot
        // satisfy the ORDER BY because status is not part of the predicate.
        DB::statement(
            'CREATE INDEX CONCURRENTLY IF NOT EXISTS "idx_game_periods_room_draw_at" '
            .'ON "game_periods" ("room_code", "draw_at" DESC)'
        );
    }

    public function down(): void
    {
        if (DB::getDriverName() === 'pgsql') {
            DB::statement('DROP INDEX CONCURRENTLY IF EXISTS "idx_game_periods_room_draw_at"');
        }
    }
};
