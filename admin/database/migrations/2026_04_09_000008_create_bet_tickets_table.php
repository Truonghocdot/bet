<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('bet_tickets', function (Blueprint $table) {
            $table->id();
            $table->string('ticket_no', 40)->unique();
            $table->foreignId('user_id')->constrained('users')->restrictOnDelete();
            $table->foreignId('wallet_id')->constrained('wallets')->restrictOnDelete();
            $table->unsignedTinyInteger('unit');
            $table->unsignedTinyInteger('game_type');
            $table->foreignId('period_id')->constrained('game_periods')->restrictOnDelete();
            $table->unsignedTinyInteger('bet_type');
            $table->decimal('stake', 20, 8);
            $table->decimal('total_odds', 14, 6);
            $table->decimal('potential_payout', 20, 8);
            $table->decimal('actual_payout', 20, 8)->nullable();
            $table->unsignedTinyInteger('status');
            $table->string('placed_ip', 45)->nullable();
            $table->string('placed_device', 100)->nullable();
            $table->timestamp('settled_at')->nullable();
            $table->timestamps();

            $table->index(['user_id', 'created_at']);
            $table->index(['period_id', 'status']);
            $table->index(['game_type', 'created_at']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('bet_tickets');
    }
};
