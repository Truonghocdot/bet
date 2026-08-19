<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('wheel_campaigns', function (Blueprint $table): void {
            $table->id();
            $table->string('name', 160);
            $table->string('status', 20)->default('draft');
            $table->timestamp('opens_at')->nullable();
            $table->timestamp('closes_at')->nullable();
            $table->unsignedSmallInteger('duration_seconds')->default(300);
            $table->unsignedSmallInteger('spin_duration_seconds')->default(5);
            $table->foreignId('created_by')->nullable()->constrained('users')->nullOnDelete();
            $table->timestamps();
            $table->index(['status', 'opens_at', 'closes_at']);
        });

        Schema::create('wheel_campaign_round_templates', function (Blueprint $table): void {
            $table->id();
            $table->foreignId('campaign_id')->constrained('wheel_campaigns')->cascadeOnDelete();
            $table->unsignedTinyInteger('round_no');
            $table->string('segment_key', 64);
            $table->string('result_label', 160);
            $table->decimal('prize_amount', 30, 8)->default(0);
            $table->timestamps();
            $table->unique(['campaign_id', 'round_no']);
        });

        Schema::create('wheel_invitations', function (Blueprint $table): void {
            $table->id();
            $table->uuid('public_id')->unique();
            $table->foreignId('campaign_id')->constrained('wheel_campaigns')->restrictOnDelete();
            $table->foreignId('user_id')->constrained('users')->restrictOnDelete();
            $table->string('status', 20)->default('draft');
            $table->timestamp('activated_at')->nullable();
            $table->timestamp('expires_at')->nullable();
            $table->timestamp('popup_seen_at')->nullable();
            $table->timestamp('revoked_at')->nullable();
            $table->foreignId('activated_by')->nullable()->constrained('users')->nullOnDelete();
            $table->foreignId('revoked_by')->nullable()->constrained('users')->nullOnDelete();
            $table->timestamps();
            $table->unique(['campaign_id', 'user_id']);
            $table->index(['user_id', 'status', 'expires_at']);
        });

        Schema::create('wheel_invitation_rounds', function (Blueprint $table): void {
            $table->id();
            $table->foreignId('invitation_id')->constrained('wheel_invitations')->cascadeOnDelete();
            $table->unsignedTinyInteger('round_no');
            $table->string('segment_key', 64);
            $table->string('result_label', 160);
            $table->decimal('prize_amount', 30, 8)->default(0);
            $table->string('status', 20)->default('pending');
            $table->timestamp('spun_at')->nullable();
            $table->timestamps();
            $table->unique(['invitation_id', 'round_no']);
        });

        Schema::create('wheel_sessions', function (Blueprint $table): void {
            $table->id();
            $table->uuid('public_id')->unique();
            $table->foreignId('invitation_id')->unique()->constrained('wheel_invitations')->restrictOnDelete();
            $table->foreignId('user_id')->constrained('users')->restrictOnDelete();
            $table->string('status', 20)->default('active');
            $table->unsignedTinyInteger('current_round')->default(1);
            $table->timestamp('started_at');
            $table->timestamp('ends_at');
            $table->timestamp('completed_at')->nullable();
            $table->timestamp('expired_at')->nullable();
            $table->unsignedInteger('version')->default(1);
            $table->timestamps();
            $table->index(['status', 'ends_at']);
            $table->index(['user_id', 'status']);
        });

        Schema::create('wheel_rewards', function (Blueprint $table): void {
            $table->id();
            $table->foreignId('session_id')->constrained('wheel_sessions')->restrictOnDelete();
            $table->foreignId('invitation_round_id')->constrained('wheel_invitation_rounds')->restrictOnDelete();
            $table->foreignId('user_id')->constrained('users')->restrictOnDelete();
            $table->unsignedTinyInteger('round_no');
            $table->unsignedTinyInteger('unit')->default(1);
            $table->decimal('amount', 30, 8);
            $table->string('status', 20)->default('pending');
            $table->string('idempotency_key', 160)->unique();
            $table->foreignId('wallet_ledger_entry_id')->nullable()->constrained('wallet_ledger_entries')->nullOnDelete();
            $table->unsignedSmallInteger('attempts')->default(0);
            $table->text('last_error')->nullable();
            $table->timestamp('paid_at')->nullable();
            $table->timestamps();
            $table->unique(['session_id', 'round_no']);
            $table->index(['status', 'created_at']);
        });

        Schema::create('wheel_outbox_events', function (Blueprint $table): void {
            $table->id();
            $table->string('topic', 190);
            $table->string('event', 80);
            $table->json('payload');
            $table->unsignedSmallInteger('attempts')->default(0);
            $table->timestamp('available_at')->nullable();
            $table->timestamp('published_at')->nullable();
            $table->text('last_error')->nullable();
            $table->timestamps();
            $table->index(['published_at', 'available_at']);
        });

        Schema::create('wheel_audit_logs', function (Blueprint $table): void {
            $table->id();
            $table->foreignId('campaign_id')->nullable()->constrained('wheel_campaigns')->nullOnDelete();
            $table->foreignId('invitation_id')->nullable()->constrained('wheel_invitations')->nullOnDelete();
            $table->foreignId('session_id')->nullable()->constrained('wheel_sessions')->nullOnDelete();
            $table->foreignId('actor_user_id')->nullable()->constrained('users')->nullOnDelete();
            $table->string('action', 80);
            $table->json('old_values')->nullable();
            $table->json('new_values')->nullable();
            $table->string('ip_address', 64)->nullable();
            $table->timestamp('created_at');
            $table->index(['action', 'created_at']);
        });

        Schema::table('chat_rooms', function (Blueprint $table): void {
            $table->foreignId('wheel_session_id')->nullable()->unique()->after('id')->constrained('wheel_sessions')->cascadeOnDelete();
            $table->timestamp('next_bot_at')->nullable()->after('enabled');
            $table->index(['enabled', 'next_bot_at']);
        });

        Schema::table('chat_bans', function (Blueprint $table): void {
            $table->foreignId('room_id')->nullable()->after('id')->constrained('chat_rooms')->cascadeOnDelete();
            $table->index(['room_id', 'user_id', 'revoked_at']);
        });

        if (DB::getDriverName() === 'pgsql') {
            DB::statement('ALTER TABLE wheel_campaigns ADD CONSTRAINT wheel_campaigns_duration_check CHECK (duration_seconds = 300 AND spin_duration_seconds = 5)');
            DB::statement('ALTER TABLE wheel_campaign_round_templates ADD CONSTRAINT wheel_campaign_round_no_check CHECK (round_no BETWEEN 1 AND 4)');
            DB::statement('ALTER TABLE wheel_invitation_rounds ADD CONSTRAINT wheel_invitation_round_no_check CHECK (round_no BETWEEN 1 AND 4)');
            DB::statement('ALTER TABLE wheel_campaign_round_templates ADD CONSTRAINT wheel_campaign_prize_nonnegative_check CHECK (prize_amount >= 0)');
            DB::statement('ALTER TABLE wheel_invitation_rounds ADD CONSTRAINT wheel_invitation_prize_nonnegative_check CHECK (prize_amount >= 0)');
            DB::statement('ALTER TABLE wheel_rewards ADD CONSTRAINT wheel_reward_amount_positive_check CHECK (amount > 0)');
        }
    }

    public function down(): void
    {
        Schema::table('chat_bans', function (Blueprint $table): void {
            $table->dropConstrainedForeignId('room_id');
        });
        Schema::table('chat_rooms', function (Blueprint $table): void {
            $table->dropIndex(['enabled', 'next_bot_at']);
            $table->dropConstrainedForeignId('wheel_session_id');
            $table->dropColumn('next_bot_at');
        });
        Schema::dropIfExists('wheel_audit_logs');
        Schema::dropIfExists('wheel_outbox_events');
        Schema::dropIfExists('wheel_rewards');
        Schema::dropIfExists('wheel_sessions');
        Schema::dropIfExists('wheel_invitation_rounds');
        Schema::dropIfExists('wheel_invitations');
        Schema::dropIfExists('wheel_campaign_round_templates');
        Schema::dropIfExists('wheel_campaigns');
    }
};
