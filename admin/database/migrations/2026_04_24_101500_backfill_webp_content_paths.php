<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Storage;

return new class extends Migration
{
    public function up(): void
    {
        $disk = Storage::disk('public');

        $this->backfillColumnToWebp(
            table: 'news_articles',
            column: 'cover_image_path',
            disk: $disk,
        );

        $this->backfillColumnToWebp(
            table: 'banners',
            column: 'image_path',
            disk: $disk,
        );
    }

    public function down(): void
    {
        // Không rollback vì file gốc có thể đã bị xóa sau khi convert sang .webp.
    }

    private function backfillColumnToWebp(string $table, string $column, $disk): void
    {
        DB::table($table)
            ->select(['id', $column])
            ->orderBy('id')
            ->chunkById(100, function ($rows) use ($table, $column, $disk): void {
                foreach ($rows as $row) {
                    $currentPath = trim((string) ($row->{$column} ?? ''));
                    if ($currentPath === '' || str_ends_with(strtolower($currentPath), '.webp')) {
                        continue;
                    }

                    $dirname = pathinfo($currentPath, PATHINFO_DIRNAME);
                    $filename = pathinfo($currentPath, PATHINFO_FILENAME);
                    if ($filename === '') {
                        continue;
                    }

                    $webpPath = ($dirname !== '' && $dirname !== '.' ? $dirname.'/' : '').$filename.'.webp';
                    if (! $disk->exists($webpPath)) {
                        continue;
                    }

                    DB::table($table)
                        ->where('id', $row->id)
                        ->update([$column => $webpPath]);
                }
            });
    }
};
