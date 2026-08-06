<?php

namespace App\Console\Commands;

use App\Services\Backup\BackupSnapshotService;
use Illuminate\Console\Command;
use Throwable;

class VerifyBackupCommand extends Command
{
    protected $signature = 'backup:verify {snapshot : Snapshot identifier shown by backup:list} {--disk= : Filesystem disk containing the snapshot}';

    protected $description = 'Download and verify every object checksum in a backup snapshot';

    public function handle(BackupSnapshotService $backups): int
    {
        try {
            $result = $backups->verify((string) $this->argument('snapshot'), $this->option('disk') ?: null);
            $this->info("Backup verified: {$result['snapshot']} ({$result['objects']} objects, {$result['bytes']} bytes)");

            return self::SUCCESS;
        } catch (Throwable $exception) {
            report($exception);
            $this->error($exception->getMessage());

            return self::FAILURE;
        }
    }
}
