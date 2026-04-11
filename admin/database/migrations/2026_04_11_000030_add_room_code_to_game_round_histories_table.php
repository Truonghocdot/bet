<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('game_round_histories', function (Blueprint $table): void {
            if (! Schema::hasColumn('game_round_histories', 'room_code')) {
                $table->string('room_code', 40)->nullable()->after('game_type');
            }
        });

        DB::statement(<<<'SQL'
            update game_round_histories
            set room_code = case lower(trim(game_type))
                when 'wingo' then 'wingo_1m'
                when 'k3' then 'k3_1m'
                when 'lottery' then 'lottery_1m'
                else 'lottery_1m'
            end
            where room_code is null or trim(room_code) = ''
        SQL);

        DB::statement('alter table game_round_histories alter column room_code set not null');
        DB::statement('create index if not exists idx_game_round_histories_room_draw_at on game_round_histories (room_code, draw_at, id)');
    }

    public function down(): void
    {
        DB::statement('drop index if exists idx_game_round_histories_room_draw_at');

        Schema::table('game_round_histories', function (Blueprint $table): void {
            if (Schema::hasColumn('game_round_histories', 'room_code')) {
                $table->dropColumn('room_code');
            }
        });
    }
};
