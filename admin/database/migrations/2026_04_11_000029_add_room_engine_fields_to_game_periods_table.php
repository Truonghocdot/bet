<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('game_periods', function (Blueprint $table): void {
            if (! Schema::hasColumn('game_periods', 'bet_lock_at')) {
                $table->timestamp('bet_lock_at')->nullable()->after('close_at');
            }
        });

        DB::statement(<<<'SQL'
            update game_periods
            set room_code = case game_type
                when 1 then 'wingo_1m'
                when 2 then 'k3_1m'
                else 'lottery_1m'
            end
            where room_code is null or trim(room_code) = ''
        SQL);

        DB::statement(<<<'SQL'
            update game_periods
            set bet_lock_at = draw_at - interval '5 seconds'
            where bet_lock_at is null
        SQL);

        DB::statement('alter table game_periods alter column room_code set not null');
        DB::statement('alter table game_periods alter column bet_lock_at set not null');

        DB::statement('alter table game_periods drop constraint if exists game_periods_game_type_period_no_unique');
        DB::statement('create unique index if not exists game_periods_room_code_period_no_unique on game_periods (room_code, period_no)');
        DB::statement('create index if not exists idx_game_periods_room_status_draw on game_periods (room_code, status, draw_at)');
        DB::statement('create index if not exists idx_game_periods_room_lock on game_periods (room_code, bet_lock_at)');
    }

    public function down(): void
    {
        DB::statement('drop index if exists idx_game_periods_room_lock');
        DB::statement('drop index if exists idx_game_periods_room_status_draw');
        DB::statement('drop index if exists game_periods_room_code_period_no_unique');
        DB::statement('create unique index if not exists game_periods_game_type_period_no_unique on game_periods (game_type, period_no)');

        DB::statement('alter table game_periods alter column room_code drop not null');
        DB::statement('alter table game_periods alter column bet_lock_at drop not null');
        Schema::table('game_periods', function (Blueprint $table): void {
            if (Schema::hasColumn('game_periods', 'bet_lock_at')) {
                $table->dropColumn('bet_lock_at');
            }
        });
    }
};
