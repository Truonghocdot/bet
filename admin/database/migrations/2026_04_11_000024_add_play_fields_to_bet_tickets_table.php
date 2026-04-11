<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('bet_tickets', function (Blueprint $table): void {
            $table->string('request_id', 64)->nullable()->unique()->after('wallet_id');
            $table->string('connection_id', 64)->nullable()->after('request_id');
            $table->decimal('total_stake', 20, 8)->nullable()->after('connection_id');
            $table->json('items')->nullable()->after('total_stake');

            $table->index(['connection_id'], 'idx_bet_tickets_connection_id');
        });
    }

    public function down(): void
    {
        Schema::table('bet_tickets', function (Blueprint $table): void {
            $table->dropIndex('idx_bet_tickets_connection_id');
            $table->dropUnique(['request_id']);
            $table->dropColumn([
                'request_id',
                'connection_id',
                'total_stake',
                'items',
            ]);
        });
    }
};
