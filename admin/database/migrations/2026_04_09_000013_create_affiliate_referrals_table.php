<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('affiliate_referrals', function (Blueprint $table) {
            $table->id();
            $table->foreignId('affiliate_profile_id')->constrained('affiliate_profiles')->restrictOnDelete();
            $table->foreignId('referrer_user_id')->constrained('users')->restrictOnDelete();
            $table->foreignId('referred_user_id')->constrained('users')->restrictOnDelete();
            $table->foreignId('affiliate_link_id')->nullable()->constrained('affiliate_links')->nullOnDelete();
            $table->foreignId('first_deposit_transaction_id')->nullable()->constrained('transactions')->nullOnDelete();
            $table->decimal('first_deposit_amount', 20, 8)->nullable();
            $table->timestamp('qualified_at')->nullable();
            $table->unsignedTinyInteger('status')->default(1);
            $table->timestamps();

            $table->unique('referred_user_id');
            $table->index(['affiliate_profile_id', 'status']);
            $table->index(['referrer_user_id', 'status']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('affiliate_referrals');
    }
};
