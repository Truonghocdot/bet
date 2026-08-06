<?php

namespace App\Console\Commands;

use App\Services\Backup\BackupSnapshotService;
use Illuminate\Console\Command;
use Throwable;

class RestoreBackupCommand extends Command
{
    protected $signature = 'backup:restore {snapshot : Snapshot identifier shown by backup:list} {--disk= : Filesystem disk containing the snapshot} {--force : Skip the production confirmation}';

    protected $description = 'Verify and restore a backup snapshot using non-destructive upserts';

    public function handle(BackupSnapshotService $backups): int
    {
        if (! $this->option('force') && ! $this->confirm('This will overwrite matching users, wallets, affiliate relationships, news, banners, and images. Continue?')) {
            $this->warn('Restore cancelled.');

            return self::SUCCESS;
        }

        try {
            $result = $backups->restore((string) $this->argument('snapshot'), $this->option('disk') ?: null);
            $this->info("Backup restored: {$result['snapshot']} ({$result['rows']} rows, {$result['assets']} assets)");
            $this->line('Rows and files created after the snapshot were not deleted.');

            return self::SUCCESS;
        } catch (Throwable $exception) {
            report($exception);
            $this->error($exception->getMessage());

            return self::FAILURE;
        }
    }
}
