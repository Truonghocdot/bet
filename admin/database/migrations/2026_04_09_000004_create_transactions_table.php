<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('transactions', function (Blueprint $table) {
            $table->id();
            $table->foreignId('user_id')->constrained('users')->restrictOnDelete();
            $table->foreignId('wallet_id')->constrained('wallets')->restrictOnDelete();
            $table->unsignedTinyInteger('unit');
            $table->unsignedTinyInteger('type');
            $table->decimal('amount', 20, 8);
            $table->decimal('fee', 20, 8)->default(0);
            $table->decimal('net_amount', 20, 8);
            $table->unsignedTinyInteger('status');
            $table->string('provider', 50)->nullable();
            $table->string('provider_txn_id', 100)->nullable();
            $table->text('reason_failed')->nullable();
            $table->foreignId('approved_by')->nullable()->constrained('users')->nullOnDelete();
            $table->timestamp('approved_at')->nullable();
            $table->timestamps();
            $table->softDeletes();

            $table->index(['user_id', 'status']);
            $table->index(['wallet_id', 'status']);
            $table->index(['provider', 'provider_txn_id']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('transactions');
    }
};
