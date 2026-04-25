<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('exchange_rate_settings', function (Blueprint $table): void {
            $table->text('popup_message')->nullable()->after('marquee_messages');
            $table->text('latest_news_popup')->nullable()->after('popup_message');
        });

        if (Schema::hasColumn('exchange_rate_settings', 'popup_pages')) {
            DB::table('exchange_rate_settings')
                ->whereNull('popup_message')
                ->whereNotNull('popup_pages')
                ->update([
                    'popup_message' => DB::raw('popup_pages'),
                ]);

            Schema::table('exchange_rate_settings', function (Blueprint $table): void {
                $table->dropColumn('popup_pages');
            });
        }
    }

    public function down(): void
    {
        Schema::table('exchange_rate_settings', function (Blueprint $table): void {
            $table->text('popup_pages')->nullable()->after('marquee_messages');
        });

        if (Schema::hasColumn('exchange_rate_settings', 'popup_message')) {
            DB::table('exchange_rate_settings')
                ->whereNull('popup_pages')
                ->whereNotNull('popup_message')
                ->update([
                    'popup_pages' => DB::raw('popup_message'),
                ]);
        }

        Schema::table('exchange_rate_settings', function (Blueprint $table): void {
            $table->dropColumn([
                'popup_message',
                'latest_news_popup',
            ]);
        });
    }
};
