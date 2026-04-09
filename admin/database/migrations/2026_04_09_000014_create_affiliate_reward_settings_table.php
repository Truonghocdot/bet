<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('affiliate_reward_settings', function (Blueprint $table) {
            $table->id();
            $table->string('name', 100);
            $table->unsignedInteger('required_qualified_referrals');
            $table->decimal('reward_amount', 20, 8);
            $table->unsignedTinyInteger('unit')->default(1);
            $table->boolean('is_active')->default(true);
            $table->timestamp('effective_from')->nullable();
            $table->timestamp('effective_to')->nullable();
            $table->string('note', 255)->nullable();
            $table->timestamps();

            $table->index(['is_active', 'unit']);
            $table->index(['required_qualified_referrals', 'unit']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('affiliate_reward_settings');
    }
};
