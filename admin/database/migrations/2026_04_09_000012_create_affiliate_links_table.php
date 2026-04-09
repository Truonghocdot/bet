<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('affiliate_links', function (Blueprint $table) {
            $table->id();
            $table->foreignId('affiliate_profile_id')->constrained('affiliate_profiles')->restrictOnDelete();
            $table->string('campaign_name', 100);
            $table->string('tracking_code', 100)->unique();
            $table->string('landing_url', 255);
            $table->unsignedTinyInteger('status')->default(1);
            $table->timestamps();

            $table->index(['affiliate_profile_id', 'status']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('affiliate_links');
    }
};
