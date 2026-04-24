<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('banners', function (Blueprint $table): void {
            $table->string('placement', 32)->default('home')->after('link_url');
            $table->index(['placement', 'is_active', 'sort_order'], 'banners_placement_active_sort_idx');
        });

        DB::table('banners')
            ->whereNull('placement')
            ->update(['placement' => 'home']);
    }

    public function down(): void
    {
        Schema::table('banners', function (Blueprint $table): void {
            $table->dropIndex('banners_placement_active_sort_idx');
            $table->dropColumn('placement');
        });
    }
};
