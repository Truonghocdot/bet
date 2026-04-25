<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('exchange_rate_settings', function (Blueprint $table): void {
            $table->text('popup_pages')->nullable()->after('marquee_messages');
        });
    }

    public function down(): void
    {
        Schema::table('exchange_rate_settings', function (Blueprint $table): void {
            $table->dropColumn('popup_pages');
        });
    }
};
