<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('bet_settlements', function (Blueprint $table) {
            $table->id();
            $table->foreignId('ticket_id')->constrained('bet_tickets')->restrictOnDelete();
            $table->foreignId('period_id')->constrained('game_periods')->restrictOnDelete();
            $table->unsignedTinyInteger('settlement_type');
            $table->unsignedTinyInteger('before_status');
            $table->unsignedTinyInteger('after_status');
            $table->decimal('payout_amount', 20, 8);
            $table->decimal('profit_loss', 20, 8);
            $table->string('note', 255)->nullable();
            $table->foreignId('settled_by')->nullable()->constrained('users')->nullOnDelete();
            $table->timestamp('created_at');

            $table->index(['ticket_id', 'created_at']);
            $table->index(['period_id', 'created_at']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('bet_settlements');
    }
};
