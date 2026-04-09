<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('account_withdrawal_infos', function (Blueprint $table) {
            $table->id();
            $table->foreignId('user_id')->constrained('users')->restrictOnDelete();
            $table->unsignedTinyInteger('unit');
            $table->string('provider_code', 50);
            $table->string('account_name', 255);
            $table->string('account_number', 255);
            $table->boolean('is_default')->default(false);
            $table->timestamps();
            $table->softDeletes();

            $table->index(['user_id', 'unit']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('account_withdrawal_infos');
    }
};
