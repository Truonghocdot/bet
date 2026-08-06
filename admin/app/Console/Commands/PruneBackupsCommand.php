<?php

namespace App\Console\Commands;

use App\Services\Backup\BackupSnapshotService;
use Illuminate\Console\Command;
use Throwable;

class PruneBackupsCommand extends Command
{
    protected $signature = 'backup:prune {--disk= : Filesystem disk containing the snapshots}';

    protected $description = 'Delete completed snapshots outside the configured retention policy';

    public function handle(BackupSnapshotService $backups): int
    {
        try {
            $deleted = $backups->prune($this->option('disk') ?: null);
            $this->info('Expired snapshots removed: '.count($deleted));

            foreach ($deleted as $snapshot) {
                $this->line($snapshot);
            }

            return self::SUCCESS;
        } catch (Throwable $exception) {
            report($exception);
            $this->error($exception->getMessage());

            return self::FAILURE;
        }
    }
}

