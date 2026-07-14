<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('exchange_rate_settings', function (Blueprint $table): void {
            $table->string('app_header_logo_path', 255)
                ->nullable()
                ->after('telegram_cskh_link');
        });
    }

    public function down(): void
    {
        Schema::table('exchange_rate_settings', function (Blueprint $table): void {
            $table->dropColumn('app_header_logo_path');
        });
    }
};
