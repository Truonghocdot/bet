<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('chat_rooms', function (Blueprint $table): void {
            $table->id();
            $table->string('code', 64)->unique();
            $table->string('name', 120);
            $table->boolean('enabled')->default(false);
            $table->timestamps();
        });

        Schema::create('chat_bot_profiles', function (Blueprint $table): void {
            $table->id();
            $table->string('display_name', 120);
            $table->string('avatar_path')->nullable();
            $table->boolean('active')->default(true);
            $table->unsignedInteger('sort_order')->default(0);
            $table->timestamps();
        });

        Schema::create('chat_user_profiles', function (Blueprint $table): void {
            $table->id();
            $table->foreignId('user_id')->unique()->constrained('users')->cascadeOnDelete();
            $table->string('display_name', 80)->unique();
            $table->timestamps();
        });

        Schema::create('chat_bot_templates', function (Blueprint $table): void {
            $table->id();
            $table->foreignId('bot_profile_id')->nullable()->constrained('chat_bot_profiles')->nullOnDelete();
            $table->string('body', 280);
            $table->string('category', 60)->default('general');
            $table->string('language', 12)->default('vi');
            $table->boolean('active')->default(true);
            $table->timestamp('last_used_at')->nullable();
            $table->unsignedInteger('usage_count')->default(0);
            $table->timestamps();
            $table->index(['active', 'category']);
        });

        Schema::create('chat_messages', function (Blueprint $table): void {
            $table->id();
            $table->foreignId('room_id')->constrained('chat_rooms')->cascadeOnDelete();
            $table->string('actor_type', 16);
            $table->foreignId('user_id')->nullable()->constrained('users')->nullOnDelete();
            $table->foreignId('bot_profile_id')->nullable()->constrained('chat_bot_profiles')->nullOnDelete();
            $table->string('display_name', 80);
            $table->string('body', 280);
            $table->unsignedTinyInteger('status')->default(1);
            $table->foreignId('moderated_by')->nullable()->constrained('users')->nullOnDelete();
            $table->timestamp('moderated_at')->nullable();
            $table->timestamps();
            $table->index(['room_id', 'status', 'id']);
            $table->index(['actor_type', 'created_at']);
            $table->index(['user_id', 'created_at']);
        });

        Schema::create('chat_bans', function (Blueprint $table): void {
            $table->id();
            $table->foreignId('user_id')->constrained('users')->cascadeOnDelete();
            $table->foreignId('created_by')->constrained('users')->restrictOnDelete();
            $table->string('reason', 255)->nullable();
            $table->timestamp('expires_at')->nullable();
            $table->timestamp('revoked_at')->nullable();
            $table->foreignId('revoked_by')->nullable()->constrained('users')->nullOnDelete();
            $table->timestamps();
            $table->index(['user_id', 'revoked_at', 'expires_at']);
        });

        Schema::create('chat_moderation_actions', function (Blueprint $table): void {
            $table->id();
            $table->foreignId('actor_user_id')->constrained('users')->restrictOnDelete();
            $table->foreignId('target_user_id')->nullable()->constrained('users')->nullOnDelete();
            $table->foreignId('message_id')->nullable()->constrained('chat_messages')->nullOnDelete();
            $table->string('action', 32);
            $table->string('reason', 255)->nullable();
            $table->json('meta')->nullable();
            $table->timestamps();
            $table->index(['action', 'created_at']);
        });

        Schema::create('chat_message_reports', function (Blueprint $table): void {
            $table->id();
            $table->foreignId('message_id')->constrained('chat_messages')->cascadeOnDelete();
            $table->foreignId('reporter_user_id')->constrained('users')->cascadeOnDelete();
            $table->string('reason', 80);
            $table->string('status', 20)->default('open');
            $table->timestamps();
            $table->unique(['message_id', 'reporter_user_id']);
        });

        DB::table('chat_rooms')->insert([
            'code' => 'global',
            'name' => 'Chat Global',
            'enabled' => true,
            'created_at' => now(),
            'updated_at' => now(),
        ]);
    }

    public function down(): void
    {
        Schema::dropIfExists('chat_message_reports');
        Schema::dropIfExists('chat_moderation_actions');
        Schema::dropIfExists('chat_bans');
        Schema::dropIfExists('chat_messages');
        Schema::dropIfExists('chat_bot_templates');
        Schema::dropIfExists('chat_user_profiles');
        Schema::dropIfExists('chat_bot_profiles');
        Schema::dropIfExists('chat_rooms');
    }
};
