<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    /**
     * Run the migrations.
     */
    public function up(): void
    {
        Schema::table('game_periods', function (Blueprint $table) {
            if (!Schema::hasColumn('game_periods', 'manual_result')) {
                // Column used for manual result intervention in Control Panel
                $table->jsonb('manual_result')->nullable();
            }
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::table('game_periods', function (Blueprint $table) {
            if (Schema::hasColumn('game_periods', 'manual_result')) {
                $table->dropColumn('manual_result');
            }
        });
    }
};
