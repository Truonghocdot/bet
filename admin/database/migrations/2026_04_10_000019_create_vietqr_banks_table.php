<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('vietqr_banks', function (Blueprint $table): void {
            $table->id();
            $table->unsignedInteger('source_id')->unique();
            $table->string('code', 50)->unique();
            $table->string('name', 255);
            $table->string('short_name', 100);
            $table->string('bin', 20)->unique();
            $table->string('logo', 255)->nullable();
            $table->boolean('transfer_supported')->default(false);
            $table->boolean('lookup_supported')->default(false);
            $table->unsignedTinyInteger('support')->nullable();
            $table->json('raw_payload')->nullable();
            $table->timestamp('synced_at')->nullable();
            $table->timestamps();

            $table->index(['short_name']);
            $table->index(['transfer_supported', 'lookup_supported']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('vietqr_banks');
    }
};
