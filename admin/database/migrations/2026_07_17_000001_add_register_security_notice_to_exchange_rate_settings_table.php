<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('exchange_rate_settings', function (Blueprint $table): void {
            $table->boolean('register_security_notice_enabled')
                ->default(false)
                ->after('notification_image_force_cancel_enabled');
            $table->text('register_security_notice_text')
                ->nullable()
                ->after('register_security_notice_enabled');
        });
    }

    public function down(): void
    {
        Schema::table('exchange_rate_settings', function (Blueprint $table): void {
            $table->dropColumn([
                'register_security_notice_enabled',
                'register_security_notice_text',
            ]);
        });
    }
};
