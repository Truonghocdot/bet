<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('telegram_chat_destinations', function (Blueprint $table): void {
            $table->id();
            $table->string('site_code', 32);
            $table->bigInteger('telegram_chat_id');
            $table->string('chat_type', 20);
            $table->string('title', 160)->nullable();
            $table->string('username', 80)->nullable();
            $table->string('bot_status', 32);
            $table->boolean('is_active')->default(false);
            $table->timestamp('discovered_at');
            $table->timestamp('last_seen_at');
            $table->timestamp('removed_at')->nullable();
            $table->text('last_error')->nullable();
            $table->timestamp('last_error_at')->nullable();
            $table->foreignId('activated_by')->nullable()->constrained('users')->nullOnDelete();
            $table->timestamp('activated_at')->nullable();
            $table->foreignId('deactivated_by')->nullable()->constrained('users')->nullOnDelete();
            $table->timestamp('deactivated_at')->nullable();
            $table->timestamps();

            $table->unique(['site_code', 'telegram_chat_id']);
            $table->index(['site_code', 'is_active', 'bot_status']);
        });

        Schema::create('telegram_chat_destination_audits', function (Blueprint $table): void {
            $table->id();
            $table->foreignId('destination_id')->constrained('telegram_chat_destinations')->cascadeOnDelete();
            $table->foreignId('actor_user_id')->nullable()->constrained('users')->nullOnDelete();
            $table->string('action', 40);
            $table->json('old_values')->nullable();
            $table->json('new_values')->nullable();
            $table->string('ip_address', 64)->nullable();
            $table->timestamp('created_at');

            $table->index(['destination_id', 'created_at']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('telegram_chat_destination_audits');
        Schema::dropIfExists('telegram_chat_destinations');
    }
};
