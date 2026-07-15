<?php

use App\Support\Media\WebpImageConverter;
use Illuminate\Database\Migrations\Migration;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        if (! Schema::hasColumn('exchange_rate_settings', 'app_header_logo_path')) {
            return;
        }

        $records = DB::table('exchange_rate_settings')
            ->select(['id', 'app_header_logo_path'])
            ->whereNotNull('app_header_logo_path')
            ->get();

        foreach ($records as $record) {
            WebpImageConverter::ensurePublicDiskWebpVariant($record->app_header_logo_path);
        }
    }

    public function down(): void
    {
        // No-op: keep generated sidecar webp files.
    }
};
