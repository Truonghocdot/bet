<?php

namespace App\Console\Commands;

use App\Services\Backup\BackupSnapshotService;
use Illuminate\Console\Command;
use Throwable;

class ListBackupsCommand extends Command
{
    protected $signature = 'backup:list {--disk= : Filesystem disk containing the snapshots}';

    protected $description = 'List completed backup snapshots';

    public function handle(BackupSnapshotService $backups): int
    {
        try {
            $rows = array_map(static function (array $manifest): array {
                return [
                    $manifest['snapshot'],
                    $manifest['created_at'],
                    array_sum(array_column($manifest['database']['tables'], 'rows')),
                    count($manifest['assets']['files']),
                ];
            }, $backups->all($this->option('disk') ?: null));

            $this->table(['Snapshot', 'Created (UTC)', 'Rows', 'Assets'], $rows);

            return self::SUCCESS;
        } catch (Throwable $exception) {
            report($exception);
            $this->error($exception->getMessage());

            return self::FAILURE;
        }
    }
}

