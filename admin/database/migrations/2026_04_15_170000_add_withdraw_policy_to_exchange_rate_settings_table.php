<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('exchange_rate_settings', function (Blueprint $table): void {
            $table->decimal('withdraw_fee_percent', 8, 4)->default(0)->after('telegram_cskh_link');
            $table->decimal('withdraw_required_bet_volume', 20, 8)->default(0)->after('withdraw_fee_percent');
            $table->unsignedInteger('withdraw_max_times_per_day')->default(3)->after('withdraw_required_bet_volume');
            $table->decimal('withdraw_min_amount', 20, 8)->default(200000)->after('withdraw_max_times_per_day');
            $table->decimal('withdraw_max_amount', 20, 8)->default(20000000)->after('withdraw_min_amount');
        });
    }

    public function down(): void
    {
        Schema::table('exchange_rate_settings', function (Blueprint $table): void {
            $table->dropColumn([
                'withdraw_fee_percent',
                'withdraw_required_bet_volume',
                'withdraw_max_times_per_day',
                'withdraw_min_amount',
                'withdraw_max_amount',
            ]);
        });
    }
};

