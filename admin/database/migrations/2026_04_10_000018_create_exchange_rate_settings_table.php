<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('exchange_rate_settings', function (Blueprint $table): void {
            $table->id();
            $table->string('code', 50)->unique();
            $table->string('base_currency', 10)->default('USDT');
            $table->string('quote_currency', 10)->default('VND');
            $table->decimal('rate', 20, 8)->default(0);
            $table->decimal('source_rate', 20, 8)->nullable();
            $table->boolean('auto_sync')->default(true);
            $table->string('source_name', 100)->nullable();
            $table->timestamp('last_synced_at')->nullable();
            $table->foreignId('updated_by')->nullable()->constrained('users')->nullOnDelete();
            $table->text('note')->nullable();
            $table->timestamps();
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('exchange_rate_settings');
    }
};
