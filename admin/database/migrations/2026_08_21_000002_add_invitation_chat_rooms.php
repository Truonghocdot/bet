<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('chat_rooms', function (Blueprint $table): void {
            $table->foreignId('wheel_invitation_id')->nullable()->after('id')->unique()->constrained('wheel_invitations')->cascadeOnDelete();
            $table->unsignedSmallInteger('bot_message_count')->default(0)->after('next_bot_at');
        });
    }

    public function down(): void
    {
        Schema::table('chat_rooms', function (Blueprint $table): void {
            $table->dropConstrainedForeignId('wheel_invitation_id');
            $table->dropColumn('bot_message_count');
        });
    }
};
