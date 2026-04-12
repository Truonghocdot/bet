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
        Schema::table('exchange_rate_settings', function (Blueprint $table) {
            $table->string('nowpayments_api_key', 255)->nullable()->after('note');
            $table->string('nowpayments_ipn_secret', 255)->nullable()->after('nowpayments_api_key');
            $table->string('nowpayments_payout_wallet', 255)->nullable()->after('nowpayments_ipn_secret');
            $table->boolean('nowpayments_sandbox')->default(false)->after('nowpayments_payout_wallet');
        });
    }

    /**
     * Reverse the migrations.
     */
    public function down(): void
    {
        Schema::table('exchange_rate_settings', function (Blueprint $table) {
            $table->dropColumn([
                'nowpayments_api_key',
                'nowpayments_ipn_secret',
                'nowpayments_payout_wallet',
                'nowpayments_sandbox',
            ]);
        });
    }
};
