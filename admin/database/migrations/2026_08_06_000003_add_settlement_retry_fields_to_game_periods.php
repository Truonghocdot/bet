<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    private const INDEX_NAME = 'idx_game_periods_settlement_retry_queue';

    public function up(): void
    {
        if (! Schema::hasTable('game_periods')) {
            return;
        }

        Schema::table('game_periods', function (Blueprint $table): void {
            if (! Schema::hasColumn('game_periods', 'settlement_attempts')) {
                $table->unsignedInteger('settlement_attempts')->default(0);
            }
            if (! Schema::hasColumn('game_periods', 'settlement_last_error')) {
                $table->text('settlement_last_error')->nullable();
            }
            if (! Schema::hasColumn('game_periods', 'settlement_next_retry_at')) {
                $table->timestamp('settlement_next_retry_at')->nullable();
            }
        });

        if (DB::getDriverName() === 'pgsql') {
            DB::statement(
                'CREATE INDEX IF NOT EXISTS "'.self::INDEX_NAME.'" '
                .'ON "game_periods" ("status", "settlement_next_retry_at", "draw_at", "id")'
            );
        }
    }

    public function down(): void
    {
        if (! Schema::hasTable('game_periods')) {
            return;
        }

        if (DB::getDriverName() === 'pgsql') {
            DB::statement('DROP INDEX IF EXISTS "'.self::INDEX_NAME.'"');
        }

        Schema::table('game_periods', function (Blueprint $table): void {
            $columns = [
                'settlement_attempts',
                'settlement_last_error',
                'settlement_next_retry_at',
            ];

            foreach ($columns as $column) {
                if (Schema::hasColumn('game_periods', $column)) {
                    $table->dropColumn($column);
                }
            }
        });
    }
};
