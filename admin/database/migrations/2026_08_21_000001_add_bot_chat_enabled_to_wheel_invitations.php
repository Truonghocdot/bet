<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('wheel_invitations', function (Blueprint $table): void {
            $table->boolean('bot_chat_enabled')->default(false)->after('status');
            $table->index(['bot_chat_enabled', 'status']);
        });
    }

    public function down(): void
    {
        Schema::table('wheel_invitations', function (Blueprint $table): void {
            $table->dropIndex(['bot_chat_enabled', 'status']);
            $table->dropColumn('bot_chat_enabled');
        });
    }
};
