<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('auth_action_limits', function (Blueprint $table) {
            $table->id();
            $table->string('scope', 50);
            $table->string('action', 50);
            $table->string('subject_key', 255);
            $table->unsignedInteger('hit_count')->default(0);
            $table->timestamp('window_started_at');
            $table->timestamp('window_ended_at');
            $table->timestamp('last_hit_at');
            $table->json('meta')->nullable();
            $table->timestamps();

            $table->index(['scope', 'action']);
            $table->index(['subject_key', 'action']);
            $table->index('window_ended_at');
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('auth_action_limits');
    }
};

