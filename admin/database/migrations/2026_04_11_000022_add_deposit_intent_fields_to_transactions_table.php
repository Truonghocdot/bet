<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('transactions', function (Blueprint $table): void {
            $table->string('client_ref', 100)->nullable()->unique()->after('wallet_id');
            $table->foreignId('receiving_account_id')
                ->nullable()
                ->constrained('payment_receiving_accounts')
                ->nullOnDelete()
                ->after('provider_txn_id');
            $table->json('meta')->nullable()->after('receiving_account_id');
            $table->index(['provider', 'provider_txn_id']);
            $table->index('receiving_account_id');
        });
    }

    public function down(): void
    {
        Schema::table('transactions', function (Blueprint $table): void {
            $table->dropIndex(['provider', 'provider_txn_id']);
            $table->dropIndex(['receiving_account_id']);
            $table->dropConstrainedForeignId('receiving_account_id');
            $table->dropColumn(['client_ref', 'meta']);
        });
    }
};
