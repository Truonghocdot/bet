<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('game_round_histories', function (Blueprint $table): void {
            $table->id();
            $table->string('game_type', 32);
            $table->string('period_no', 64);
            $table->string('result', 64);
            $table->string('big_small', 32);
            $table->string('color', 32);
            $table->timestampTz('draw_at');
            $table->string('status', 24)->default('DRAWN');
            $table->timestampsTz();

            $table->index(['game_type', 'draw_at', 'id'], 'idx_game_round_histories_game_type_draw_at');
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('game_round_histories');
    }
};
