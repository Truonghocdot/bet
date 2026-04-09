<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('game_periods', function (Blueprint $table) {
            $table->id();
            $table->unsignedTinyInteger('game_type');
            $table->string('period_no', 50);
            $table->string('room_code', 30)->nullable();
            $table->timestamp('open_at');
            $table->timestamp('close_at');
            $table->timestamp('draw_at');
            $table->timestamp('settled_at')->nullable();
            $table->unsignedTinyInteger('status');
            $table->unsignedTinyInteger('draw_source')->nullable();
            $table->json('result_payload')->nullable();
            $table->string('result_hash', 255)->nullable();
            $table->timestamps();

            $table->unique(['game_type', 'period_no']);
            $table->index(['status', 'draw_at']);
            $table->index(['game_type', 'close_at']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('game_periods');
    }
};
