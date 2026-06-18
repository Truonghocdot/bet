<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('fake_deposit_transactions', function (Blueprint $table): void {
            $table->id();
            $table->unsignedInteger('source_index')->default(0)->index();
            $table->string('masked_code', 32);
            $table->string('masked_phone', 32);
            $table->string('status_label', 64);
            $table->json('meta')->nullable();
            $table->timestamps();
        });

        Schema::create('fake_withdraw_transactions', function (Blueprint $table): void {
            $table->id();
            $table->unsignedInteger('source_index')->default(0)->index();
            $table->string('masked_code', 32);
            $table->string('masked_phone', 32);
            $table->string('status_label', 64);
            $table->json('meta')->nullable();
            $table->timestamps();
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('fake_withdraw_transactions');
        Schema::dropIfExists('fake_deposit_transactions');
    }
};
