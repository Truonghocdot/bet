<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('affiliate_reward_logs', function (Blueprint $table) {
            $table->id();
            $table->foreignId('affiliate_profile_id')->constrained('affiliate_profiles')->restrictOnDelete();
            $table->foreignId('referrer_user_id')->constrained('users')->restrictOnDelete();
            $table->foreignId('setting_id')->constrained('affiliate_reward_settings')->restrictOnDelete();
            $table->unsignedInteger('required_qualified_referrals');
            $table->unsignedInteger('actual_qualified_referrals');
            $table->decimal('reward_amount', 20, 8);
            $table->unsignedTinyInteger('unit');
            $table->unsignedTinyInteger('status')->default(1);
            $table->foreignId('wallet_ledger_entry_id')->nullable()->constrained('wallet_ledger_entries')->nullOnDelete();
            $table->timestamp('granted_at')->nullable();
            $table->timestamps();

            $table->unique(['affiliate_profile_id', 'setting_id']);
            $table->index(['referrer_user_id', 'status']);
            $table->index(['affiliate_profile_id', 'status']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('affiliate_reward_logs');
    }
};
