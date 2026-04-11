<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('game_rooms', function (Blueprint $table): void {
            $table->id();
            $table->string('code', 40)->unique();
            $table->unsignedTinyInteger('game_type');
            $table->unsignedInteger('duration_seconds');
            $table->unsignedSmallInteger('bet_cutoff_seconds')->default(5);
            $table->unsignedTinyInteger('status')->default(1);
            $table->unsignedInteger('sort_order')->default(0);
            $table->timestamps();

            $table->index(['game_type', 'status', 'sort_order']);
            $table->index(['status', 'sort_order']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('game_rooms');
    }
};
