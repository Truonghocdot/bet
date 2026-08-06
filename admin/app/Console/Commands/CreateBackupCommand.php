<?php

namespace App\Console\Commands;

use App\Services\Backup\BackupSnapshotService;
use Illuminate\Console\Command;
use Throwable;

class CreateBackupCommand extends Command
{
    protected $signature = 'backup:create {--disk= : Filesystem disk used for the snapshot} {--prune : Prune expired snapshots after a successful backup}';

    protected $description = 'Back up users, wallets, affiliate relationships, news, banners, and their images';

    public function handle(BackupSnapshotService $backups): int
    {
        try {
            $manifest = $backups->create($this->option('disk') ?: null);
            $rows = array_sum(array_column($manifest['database']['tables'], 'rows'));
            $this->info("Backup completed: {$manifest['snapshot']} ({$rows} rows, ".count($manifest['assets']['files']).' assets)');

            if ($this->option('prune')) {
                $deleted = $backups->prune($this->option('disk') ?: null);
                $this->line('Expired snapshots removed: '.count($deleted));
            }

            return self::SUCCESS;
        } catch (Throwable $exception) {
            report($exception);
            $this->error($exception->getMessage());

            return self::FAILURE;
        }
    }
}
