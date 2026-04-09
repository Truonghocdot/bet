<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('payment_receiving_accounts', function (Blueprint $table) {
            $table->id();
            $table->string('code', 50)->unique();
            $table->string('name', 100);
            $table->unsignedTinyInteger('type');
            $table->unsignedTinyInteger('unit');
            $table->string('provider_code', 50)->nullable();
            $table->string('account_name', 255)->nullable();
            $table->string('account_number', 255)->nullable();
            $table->string('wallet_address', 255)->nullable();
            $table->string('network', 50)->nullable();
            $table->string('qr_code_path', 255)->nullable();
            $table->text('instructions')->nullable();
            $table->unsignedTinyInteger('status')->default(1);
            $table->boolean('is_default')->default(false);
            $table->unsignedInteger('sort_order')->default(0);
            $table->timestamps();
            $table->softDeletes();

            $table->index(['type', 'unit', 'status']);
            $table->index(['is_default', 'status']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('payment_receiving_accounts');
    }
};
