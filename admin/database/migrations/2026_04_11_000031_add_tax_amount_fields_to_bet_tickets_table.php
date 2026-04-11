<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('bet_tickets', function (Blueprint $table): void {
            $table->decimal('original_amount', 20, 8)->nullable()->after('stake');
            $table->decimal('tax_amount', 20, 8)->nullable()->after('original_amount');
            $table->decimal('net_amount', 20, 8)->nullable()->after('tax_amount');
        });

        DB::statement('update bet_tickets set original_amount = coalesce(total_stake, stake), tax_amount = 0, net_amount = coalesce(total_stake, stake) where original_amount is null');
    }

    public function down(): void
    {
        Schema::table('bet_tickets', function (Blueprint $table): void {
            $table->dropColumn([
                'original_amount',
                'tax_amount',
                'net_amount',
            ]);
        });
    }
};
