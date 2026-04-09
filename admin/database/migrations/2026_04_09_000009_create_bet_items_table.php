<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('bet_items', function (Blueprint $table) {
            $table->id();
            $table->foreignId('ticket_id')->constrained('bet_tickets')->restrictOnDelete();
            $table->foreignId('period_id')->constrained('game_periods')->restrictOnDelete();
            $table->unsignedTinyInteger('option_type');
            $table->string('option_key', 100);
            $table->string('option_label', 150);
            $table->decimal('odds_at_placement', 12, 4);
            $table->decimal('stake', 20, 8);
            $table->unsignedTinyInteger('result');
            $table->decimal('payout_amount', 20, 8)->nullable();
            $table->json('result_payload')->nullable();
            $table->timestamp('settled_at')->nullable();
            $table->timestamps();

            $table->index(['ticket_id']);
            $table->index(['period_id', 'result']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('bet_items');
    }
};
