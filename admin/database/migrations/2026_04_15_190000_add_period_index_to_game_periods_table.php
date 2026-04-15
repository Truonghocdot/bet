<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('game_periods', function (Blueprint $table) {
            $table->unsignedBigInteger('period_index')->nullable()->after('period_no');
        });

        DB::statement("
            update game_periods
            set period_index = coalesce(
                nullif(regexp_replace(period_no, '^.*_(\\d+)$', '\\1'), period_no)::bigint,
                extract(epoch from draw_at)::bigint
            )
            where period_index is null
        ");

        Schema::table('game_periods', function (Blueprint $table) {
            $table->unique(['room_code', 'period_index'], 'game_periods_room_code_period_index_unique');
        });
    }

    public function down(): void
    {
        Schema::table('game_periods', function (Blueprint $table) {
            $table->dropUnique('game_periods_room_code_period_index_unique');
            $table->dropColumn('period_index');
        });
    }
};

