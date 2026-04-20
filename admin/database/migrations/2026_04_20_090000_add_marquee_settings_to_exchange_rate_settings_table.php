<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('exchange_rate_settings', function (Blueprint $table): void {
            $table->boolean('marquee_enabled')->default(true)->after('telegram_cskh_link');
            $table->text('marquee_messages')->nullable()->after('marquee_enabled');
        });
    }

    public function down(): void
    {
        Schema::table('exchange_rate_settings', function (Blueprint $table): void {
            $table->dropColumn([
                'marquee_enabled',
                'marquee_messages',
            ]);
        });
    }
};
