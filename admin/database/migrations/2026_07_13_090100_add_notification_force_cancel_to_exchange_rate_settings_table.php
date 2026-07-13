<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('exchange_rate_settings', function (Blueprint $table): void {
            $table->boolean('notification_image_force_cancel_enabled')->default(false)->after('fake_finance_feed_enabled');
        });
    }

    public function down(): void
    {
        Schema::table('exchange_rate_settings', function (Blueprint $table): void {
            $table->dropColumn('notification_image_force_cancel_enabled');
        });
    }
};
