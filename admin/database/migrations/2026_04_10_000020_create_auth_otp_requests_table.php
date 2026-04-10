<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('auth_otp_requests', function (Blueprint $table) {
            $table->id();
            $table->foreignId('user_id')->nullable()->constrained('users')->nullOnDelete();
            $table->unsignedTinyInteger('channel');
            $table->unsignedTinyInteger('purpose');
            $table->string('target', 255);
            $table->string('otp_hash', 255);
            $table->string('otp_last4', 10);
            $table->string('request_token', 100)->unique();
            $table->unsignedInteger('attempt_count')->default(0);
            $table->unsignedInteger('max_attempts')->default(5);
            $table->timestamp('expires_at');
            $table->timestamp('verified_at')->nullable();
            $table->timestamp('used_at')->nullable();
            $table->timestamp('locked_at')->nullable();
            $table->timestamp('sent_at')->nullable();
            $table->unsignedTinyInteger('status')->default(1);
            $table->json('meta')->nullable();
            $table->timestamps();

            $table->index(['user_id', 'purpose', 'status']);
            $table->index(['channel', 'target', 'purpose', 'status']);
            $table->index(['expires_at', 'status']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('auth_otp_requests');
    }
};

