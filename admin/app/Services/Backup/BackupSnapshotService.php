<?php

namespace App\Services\Backup;

use Carbon\CarbonImmutable;
use Illuminate\Database\Connection;
use Illuminate\Filesystem\AwsS3V3Adapter;
use Illuminate\Filesystem\FilesystemAdapter;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\File;
use Illuminate\Support\Facades\Storage;
use Illuminate\Support\Str;
use JsonException;
use RuntimeException;
use Throwable;

class BackupSnapshotService
{
    private const FORMAT = 'site-backup-v2';

    private const LEGACY_FORMAT = 'site-backup-v1';

    private const LEGACY_TABLES = [
        'users',
        'wallets',
        'wallet_ledger_entries',
        'news_articles',
    ];

    /**
     * @return array<string, mixed>
     */
    public function create(?string $diskName = null): array
    {
        $diskName ??= (string) config('backup.disk');
        $disk = Storage::disk($diskName);
        $snapshot = now('UTC')->format('Y/m/d/His').'-'.Str::lower((string) Str::uuid());
        $basePath = $this->snapshotPath($snapshot);
        $temporaryDirectory = storage_path('app/private/backup-tmp/'.Str::uuid());

        File::ensureDirectoryExists($temporaryDirectory, 0700, true);

        try {
            $tables = $this->exportDatabase($temporaryDirectory);

            foreach ($tables as &$table) {
                $this->uploadLocalFile(
                    $disk,
                    $temporaryDirectory.'/'.$table['local_file'],
                    $basePath.'/'.$table['path'],
                );
                unset($table['local_file']);
            }
            unset($table);

            $assets = $this->uploadAssets($disk, $basePath);
            $manifest = [
                'format' => self::FORMAT,
                'complete' => true,
                'site' => $this->site(),
                'snapshot' => $snapshot,
                'created_at' => now('UTC')->toIso8601String(),
                'application' => [
                    'name' => (string) config('app.name'),
                    'environment' => (string) app()->environment(),
                ],
                'database' => [
                    'driver' => DB::connection()->getDriverName(),
                    'tables' => $tables,
                ],
                'assets' => [
                    'source_disk' => (string) config('backup.asset_disk'),
                    'files' => $assets,
                ],
            ];

            $encodedManifest = json_encode(
                $manifest,
                JSON_PRETTY_PRINT | JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE | JSON_THROW_ON_ERROR,
            );

            $this->uploadString(
                $disk,
                $this->manifestPath($snapshot),
                $encodedManifest,
                'application/json',
            );

            return $manifest;
        } catch (Throwable $exception) {
            try {
                $disk->deleteDirectory($basePath);
                $disk->delete($this->manifestPath($snapshot));
            } catch (Throwable) {
                // The missing manifest keeps an incomplete upload from being treated as a backup.
            }

            throw $exception;
        } finally {
            File::deleteDirectory($temporaryDirectory);
        }
    }

    /**
     * @return list<array<string, mixed>>
     */
    public function all(?string $diskName = null): array
    {
        $disk = Storage::disk($diskName ?? (string) config('backup.disk'));
        $catalogRoot = $this->rootPath().'/_manifests';
        $snapshots = [];

        foreach ($disk->allFiles($catalogRoot) as $path) {
            if (! str_ends_with($path, '.json')) {
                continue;
            }

            try {
                $snapshot = Str::beforeLast(Str::after($path, $catalogRoot.'/'), '.json');
                $manifest = $this->decodeManifest($disk->get($path));
                $this->validateManifest($manifest, $snapshot);
                $snapshots[] = $manifest;
            } catch (Throwable) {
                continue;
            }
        }

        usort(
            $snapshots,
            static fn (array $left, array $right): int => strcmp((string) $right['created_at'], (string) $left['created_at']),
        );

        return $snapshots;
    }

    /**
     * @return array{snapshot: string, objects: int, bytes: int}
     */
    public function verify(string $snapshot, ?string $diskName = null): array
    {
        $disk = Storage::disk($diskName ?? (string) config('backup.disk'));
        $manifest = $this->manifest($disk, $snapshot);
        $objects = 0;
        $bytes = 0;

        foreach ($this->manifestObjects($manifest) as $object) {
            $path = $this->snapshotPath($snapshot).'/'.$object['path'];
            $actual = $this->hashRemoteObject($disk, $path);

            if (! hash_equals((string) $object['sha256'], $actual['sha256'])) {
                throw new RuntimeException("Backup checksum mismatch: {$object['path']}");
            }

            if ((int) $object['bytes'] !== $actual['bytes']) {
                throw new RuntimeException("Backup size mismatch: {$object['path']}");
            }

            $objects++;
            $bytes += $actual['bytes'];
        }

        return ['snapshot' => $snapshot, 'objects' => $objects, 'bytes' => $bytes];
    }

    /**
     * Restore is an upsert. It does not remove records or files created after the snapshot.
     *
     * @return array{snapshot: string, rows: int, assets: int}
     */
    public function restore(string $snapshot, ?string $diskName = null): array
    {
        $disk = Storage::disk($diskName ?? (string) config('backup.disk'));
        $manifest = $this->manifest($disk, $snapshot);
        $this->verify($snapshot, $diskName);
        $temporaryDirectory = storage_path('app/private/backup-restore/'.Str::uuid());
        $tableFiles = [];

        File::ensureDirectoryExists($temporaryDirectory, 0700, true);

        try {
            foreach ($manifest['database']['tables'] as $table) {
                $localPath = $temporaryDirectory.'/'.basename((string) $table['path']);
                $this->downloadToLocalFile(
                    $disk,
                    $this->snapshotPath($snapshot).'/'.$table['path'],
                    $localPath,
                );
                $tableFiles[(string) $table['name']] = $localPath;
            }

            $assetCount = $this->restoreAssets($disk, $snapshot, $manifest['assets']['files']);
            $rowCount = DB::connection()->transaction(function () use ($tableFiles): int {
                $restored = 0;

                foreach ((array) config('backup.tables', []) as $table) {
                    if (isset($tableFiles[$table])) {
                        $restored += $this->restoreTable((string) $table, $tableFiles[$table]);
                    }
                }

                $this->resetPostgresSequences(array_keys($tableFiles));

                return $restored;
            });

            return ['snapshot' => $snapshot, 'rows' => $rowCount, 'assets' => $assetCount];
        } finally {
            File::deleteDirectory($temporaryDirectory);
        }
    }

    /**
     * @return list<string>
     */
    public function prune(?string $diskName = null): array
    {
        $diskName ??= (string) config('backup.disk');
        $disk = Storage::disk($diskName);
        $snapshots = $this->all($diskName);
        $keepMinimum = max(0, (int) config('backup.keep_minimum', 7));
        $cutoff = CarbonImmutable::now('UTC')->subDays(max(1, (int) config('backup.retention_days', 30)));
        $deleted = [];

        foreach ($snapshots as $index => $manifest) {
            if ($index < $keepMinimum || CarbonImmutable::parse($manifest['created_at'])->greaterThanOrEqualTo($cutoff)) {
                continue;
            }

            $snapshot = (string) $manifest['snapshot'];
            $disk->deleteDirectory($this->snapshotPath($snapshot));
            $disk->delete($this->manifestPath($snapshot));
            $deleted[] = $snapshot;
        }

        return $deleted;
    }

    /**
     * @return list<array<string, mixed>>
     */
    private function exportDatabase(string $temporaryDirectory): array
    {
        $connection = DB::connection();
        $driver = $connection->getDriverName();

        if (in_array($driver, ['mysql', 'mariadb'], true)) {
            $connection->statement('SET TRANSACTION ISOLATION LEVEL REPEATABLE READ');
        }

        $connection->beginTransaction();

        try {
            if ($driver === 'pgsql') {
                $connection->statement('SET TRANSACTION ISOLATION LEVEL REPEATABLE READ');
            }

            $exports = [];
            foreach ((array) config('backup.tables', []) as $table) {
                $exports[] = $this->exportTable($connection, (string) $table, $temporaryDirectory);
            }
            $connection->commit();

            return $exports;
        } catch (Throwable $exception) {
            if ($connection->transactionLevel() > 0) {
                $connection->rollBack();
            }

            throw $exception;
        }
    }

    /**
     * @return array<string, mixed>
     */
    private function exportTable(Connection $connection, string $table, string $temporaryDirectory): array
    {
        $schema = $connection->getSchemaBuilder();
        if (! $schema->hasTable($table)) {
            throw new RuntimeException("Backup table does not exist: {$table}");
        }

        $columns = $schema->getColumnListing($table);
        $fileName = $table.'.jsonl.gz';
        $localPath = $temporaryDirectory.'/'.$fileName;
        $stream = gzopen($localPath, 'wb9');

        if ($stream === false) {
            throw new RuntimeException("Unable to create temporary backup for table {$table}.");
        }

        $rows = 0;

        try {
            foreach ($connection->table($table)->orderBy('id')->cursor() as $row) {
                $encoded = json_encode(
                    (array) $row,
                    JSON_UNESCAPED_SLASHES | JSON_UNESCAPED_UNICODE | JSON_PRESERVE_ZERO_FRACTION | JSON_THROW_ON_ERROR,
                )."\n";

                if (gzwrite($stream, $encoded) !== strlen($encoded)) {
                    throw new RuntimeException("Unable to write temporary backup for table {$table}.");
                }

                $rows++;
            }
        } finally {
            gzclose($stream);
        }

        $sha256 = hash_file('sha256', $localPath);
        if ($sha256 === false) {
            throw new RuntimeException("Unable to checksum temporary backup for table {$table}.");
        }

        return [
            'name' => $table,
            'path' => 'database/'.$fileName,
            'local_file' => $fileName,
            'columns' => $columns,
            'rows' => $rows,
            'bytes' => File::size($localPath),
            'sha256' => $sha256,
        ];
    }

    /**
     * @return list<array{source_path: string, path: string, bytes: int, sha256: string}>
     */
    private function uploadAssets(FilesystemAdapter $destination, string $basePath): array
    {
        $source = Storage::disk((string) config('backup.asset_disk'));
        $paths = [];

        foreach ((array) config('backup.asset_directories', []) as $directory) {
            foreach ($source->allFiles(trim((string) $directory, '/')) as $path) {
                $paths[$path] = true;
            }
        }

        $paths = array_keys($paths);
        sort($paths);
        $assets = [];

        foreach ($paths as $sourcePath) {
            $sourceStream = $source->readStream($sourcePath);
            if (! is_resource($sourceStream)) {
                throw new RuntimeException("Unable to read backup asset: {$sourcePath}");
            }

            $temporaryStream = tmpfile();
            if ($temporaryStream === false) {
                fclose($sourceStream);
                throw new RuntimeException('Unable to create a temporary asset stream.');
            }

            $hash = hash_init('sha256');
            $bytes = 0;

            try {
                while (! feof($sourceStream)) {
                    $chunk = fread($sourceStream, 1024 * 1024);
                    if ($chunk === false) {
                        throw new RuntimeException("Unable to read backup asset: {$sourcePath}");
                    }
                    if ($chunk === '') {
                        continue;
                    }

                    hash_update($hash, $chunk);
                    $bytes += strlen($chunk);
                    if (fwrite($temporaryStream, $chunk) !== strlen($chunk)) {
                        throw new RuntimeException("Unable to spool backup asset: {$sourcePath}");
                    }
                }

                rewind($temporaryStream);
                $objectPath = 'assets/'.$sourcePath;
                $this->uploadBackupStream($destination, $basePath.'/'.$objectPath, $temporaryStream);

                $assets[] = [
                    'source_path' => $sourcePath,
                    'path' => $objectPath,
                    'bytes' => $bytes,
                    'sha256' => hash_final($hash),
                ];
            } finally {
                fclose($sourceStream);
                fclose($temporaryStream);
            }
        }

        return $assets;
    }

    private function uploadLocalFile(FilesystemAdapter $disk, string $localPath, string $destinationPath): void
    {
        $stream = fopen($localPath, 'rb');
        if ($stream === false) {
            throw new RuntimeException("Unable to read temporary backup file: {$localPath}");
        }

        try {
            $this->uploadBackupStream($disk, $destinationPath, $stream);
        } finally {
            fclose($stream);
        }
    }

    private function uploadString(
        FilesystemAdapter $disk,
        string $destinationPath,
        string $contents,
        string $contentType,
    ): void {
        $stream = fopen('php://temp', 'w+b');
        if ($stream === false) {
            throw new RuntimeException('Unable to create a temporary upload stream.');
        }

        try {
            if (fwrite($stream, $contents) !== strlen($contents)) {
                throw new RuntimeException("Unable to stage backup object: {$destinationPath}");
            }

            rewind($stream);
            $this->uploadBackupStream($disk, $destinationPath, $stream, $contentType);
        } finally {
            fclose($stream);
        }
    }

    /**
     * R2 does not implement S3 ACL headers. Flysystem always supplies a private
     * ACL, so R2 writes use the underlying client with a null ACL instead.
     *
     * @param  resource  $stream
     */
    private function uploadBackupStream(
        FilesystemAdapter $disk,
        string $destinationPath,
        $stream,
        ?string $contentType = null,
    ): void {
        if ($disk instanceof AwsS3V3Adapter
            && str_contains((string) ($disk->getConfig()['endpoint'] ?? ''), 'r2.cloudflarestorage.com')) {
            $config = $disk->getConfig();
            $root = trim((string) ($config['root'] ?? ''), '/');
            $key = $root === '' ? $destinationPath : $root.'/'.$destinationPath;
            $options = $contentType === null ? [] : ['params' => ['ContentType' => $contentType]];

            $disk->getClient()->upload((string) $config['bucket'], $key, $stream, null, $options);

            return;
        }

        $options = $contentType === null ? [] : ['mimetype' => $contentType];
        if (! $disk->writeStream($destinationPath, $stream, $options)) {
            throw new RuntimeException("Unable to upload backup object: {$destinationPath}");
        }
    }

    private function downloadToLocalFile(FilesystemAdapter $disk, string $sourcePath, string $localPath): void
    {
        $source = $disk->readStream($sourcePath);
        $destination = fopen($localPath, 'wb');

        if (! is_resource($source) || $destination === false) {
            is_resource($source) && fclose($source);
            is_resource($destination) && fclose($destination);
            throw new RuntimeException("Unable to stage backup object: {$sourcePath}");
        }

        try {
            if (stream_copy_to_stream($source, $destination) === false) {
                throw new RuntimeException("Unable to stage backup object: {$sourcePath}");
            }
        } finally {
            fclose($source);
            fclose($destination);
        }
    }

    /**
     * @param  list<array<string, mixed>>  $assets
     */
    private function restoreAssets(FilesystemAdapter $source, string $snapshot, array $assets): int
    {
        $destination = Storage::disk((string) config('backup.asset_disk'));
        $restored = 0;

        foreach ($assets as $asset) {
            $sourcePath = (string) $asset['source_path'];
            $this->assertSafeRelativePath($sourcePath);
            $stream = $source->readStream($this->snapshotPath($snapshot).'/'.$asset['path']);

            if (! is_resource($stream)) {
                throw new RuntimeException("Unable to read backup asset: {$sourcePath}");
            }

            try {
                if (! $destination->writeStream($sourcePath, $stream, ['visibility' => 'public'])) {
                    throw new RuntimeException("Unable to restore backup asset: {$sourcePath}");
                }
            } finally {
                fclose($stream);
            }

            $restored++;
        }

        return $restored;
    }

    private function restoreTable(string $table, string $localPath): int
    {
        $connection = DB::connection();
        $schema = $connection->getSchemaBuilder();
        if (! $schema->hasTable($table)) {
            throw new RuntimeException("Restore table does not exist: {$table}");
        }

        $allowedColumns = array_flip($schema->getColumnListing($table));
        $stream = gzopen($localPath, 'rb');
        if ($stream === false) {
            throw new RuntimeException("Unable to read staged table backup: {$table}");
        }

        $count = 0;
        $batch = [];

        try {
            while (($line = gzgets($stream)) !== false) {
                $row = json_decode($line, true, 512, JSON_THROW_ON_ERROR);
                if (! is_array($row) || ! array_key_exists('id', $row)) {
                    throw new RuntimeException("Invalid row in table backup: {$table}");
                }

                $batch[] = array_intersect_key($row, $allowedColumns);
                if (count($batch) >= 250) {
                    $this->upsertBatch($connection, $table, $batch);
                    $count += count($batch);
                    $batch = [];
                }
            }

            if ($batch !== []) {
                $this->upsertBatch($connection, $table, $batch);
                $count += count($batch);
            }
        } catch (JsonException $exception) {
            throw new RuntimeException("Invalid JSON in table backup: {$table}", previous: $exception);
        } finally {
            gzclose($stream);
        }

        return $count;
    }

    /**
     * @param  list<array<string, mixed>>  $rows
     */
    private function upsertBatch(Connection $connection, string $table, array $rows): void
    {
        $updateColumns = array_values(array_diff(array_keys($rows[0]), ['id']));
        $connection->table($table)->upsert($rows, ['id'], $updateColumns);
    }

    /**
     * PostgreSQL does not advance a sequence when rows are restored with explicit IDs.
     *
     * @param  list<string>  $tables
     */
    private function resetPostgresSequences(array $tables): void
    {
        $connection = DB::connection();
        if ($connection->getDriverName() !== 'pgsql') {
            return;
        }

        foreach ($tables as $table) {
            if (preg_match('/^[A-Za-z_][A-Za-z0-9_]*$/', $table) !== 1) {
                throw new RuntimeException("Unsafe table name in backup configuration: {$table}");
            }

            $sequence = $connection->scalar("select pg_get_serial_sequence(?, 'id')", [$table]);
            if (! is_string($sequence) || $sequence === '') {
                continue;
            }

            $wrappedTable = $connection->getQueryGrammar()->wrapTable($table);
            $wrappedId = $connection->getQueryGrammar()->wrap('id');
            $connection->statement(
                "select setval(cast(? as regclass), greatest(coalesce(max({$wrappedId}), 1), 1), count(*) > 0) from {$wrappedTable}",
                [$sequence],
            );
        }
    }

    /**
     * @return array<string, mixed>
     */
    private function manifest(FilesystemAdapter $disk, string $snapshot): array
    {
        $this->assertSnapshot($snapshot);
        $path = $this->manifestPath($snapshot);

        if (! $disk->exists($path)) {
            throw new RuntimeException("Backup snapshot not found: {$snapshot}");
        }

        $manifest = $this->decodeManifest($disk->get($path));
        $this->validateManifest($manifest, $snapshot);

        return $manifest;
    }

    /**
     * @return array<string, mixed>
     */
    private function decodeManifest(string $contents): array
    {
        try {
            $manifest = json_decode($contents, true, 512, JSON_THROW_ON_ERROR);
        } catch (JsonException $exception) {
            throw new RuntimeException('Backup manifest contains invalid JSON.', previous: $exception);
        }

        if (! is_array($manifest)) {
            throw new RuntimeException('Backup manifest must be a JSON object.');
        }

        return $manifest;
    }

    /**
     * @param  array<string, mixed>  $manifest
     */
    private function validateManifest(array $manifest, ?string $expectedSnapshot = null): void
    {
        $format = $manifest['format'] ?? null;
        if (! in_array($format, [self::FORMAT, self::LEGACY_FORMAT], true) || ($manifest['complete'] ?? false) !== true) {
            throw new RuntimeException('Unsupported or incomplete backup manifest.');
        }

        if (($manifest['site'] ?? null) !== $this->site()) {
            throw new RuntimeException('Backup belongs to a different site.');
        }

        $snapshot = (string) ($manifest['snapshot'] ?? '');
        $this->assertSnapshot($snapshot);
        if ($expectedSnapshot !== null && $snapshot !== $expectedSnapshot) {
            throw new RuntimeException('Backup manifest snapshot does not match its object path.');
        }

        if (! isset($manifest['database']['tables'], $manifest['assets']['files'])) {
            throw new RuntimeException('Backup manifest is missing database or asset metadata.');
        }

        if (! is_array($manifest['database']['tables']) || ! is_array($manifest['assets']['files'])) {
            throw new RuntimeException('Backup manifest object lists must be arrays.');
        }

        try {
            CarbonImmutable::parse((string) ($manifest['created_at'] ?? ''));
        } catch (Throwable $exception) {
            throw new RuntimeException('Backup manifest has an invalid creation time.', previous: $exception);
        }

        $expectedTables = $format === self::LEGACY_FORMAT
            ? self::LEGACY_TABLES
            : array_values(array_map('strval', (array) config('backup.tables', [])));
        $actualTables = array_map(
            static fn (mixed $table): string => is_array($table) ? (string) ($table['name'] ?? '') : '',
            $manifest['database']['tables'],
        );
        sort($expectedTables);
        sort($actualTables);

        if ($actualTables !== $expectedTables) {
            throw new RuntimeException('Backup manifest does not contain the configured table set.');
        }

        foreach ($this->manifestObjects($manifest) as $object) {
            if (! is_array($object) || ! isset($object['path'], $object['bytes'], $object['sha256'])) {
                throw new RuntimeException('Backup manifest contains incomplete object metadata.');
            }
            $this->assertSafeRelativePath((string) $object['path']);

            if ((int) $object['bytes'] < 0 || preg_match('/^[0-9a-f]{64}$/', (string) $object['sha256']) !== 1) {
                throw new RuntimeException('Backup manifest contains invalid checksum metadata.');
            }
        }

        foreach ($manifest['assets']['files'] as $asset) {
            if (! is_array($asset) || ! isset($asset['source_path'])) {
                throw new RuntimeException('Backup manifest contains incomplete asset metadata.');
            }

            $sourcePath = (string) $asset['source_path'];
            $this->assertSafeRelativePath($sourcePath);

            $allowed = collect((array) config('backup.asset_directories', []))->contains(
                static fn (mixed $directory): bool => str_starts_with($sourcePath, trim((string) $directory, '/').'/'),
            );

            if (! $allowed) {
                throw new RuntimeException("Backup asset is outside the configured directories: {$sourcePath}");
            }
        }
    }

    /**
     * @param  array<string, mixed>  $manifest
     * @return list<array<string, mixed>>
     */
    private function manifestObjects(array $manifest): array
    {
        return array_merge($manifest['database']['tables'], $manifest['assets']['files']);
    }

    /**
     * @return array{bytes: int, sha256: string}
     */
    private function hashRemoteObject(FilesystemAdapter $disk, string $path): array
    {
        $stream = $disk->readStream($path);
        if (! is_resource($stream)) {
            throw new RuntimeException("Unable to read backup object: {$path}");
        }

        $hash = hash_init('sha256');
        $bytes = 0;

        try {
            while (! feof($stream)) {
                $chunk = fread($stream, 1024 * 1024);
                if ($chunk === false) {
                    throw new RuntimeException("Unable to verify backup object: {$path}");
                }
                if ($chunk === '') {
                    continue;
                }

                hash_update($hash, $chunk);
                $bytes += strlen($chunk);
            }
        } finally {
            fclose($stream);
        }

        return ['bytes' => $bytes, 'sha256' => hash_final($hash)];
    }

    private function rootPath(): string
    {
        $prefix = trim((string) config('backup.prefix', 'backups'), '/');

        return $prefix === '' ? $this->site() : $prefix.'/'.$this->site();
    }

    private function snapshotPath(string $snapshot): string
    {
        $this->assertSnapshot($snapshot);

        return $this->rootPath().'/'.$snapshot;
    }

    private function manifestPath(string $snapshot): string
    {
        $this->assertSnapshot($snapshot);

        return $this->rootPath().'/_manifests/'.$snapshot.'.json';
    }

    private function site(): string
    {
        $site = trim((string) config('backup.site'));
        if ($site === '' || preg_match('/^[A-Za-z0-9][A-Za-z0-9._-]*$/', $site) !== 1) {
            throw new RuntimeException('BACKUP_SITE is required and may only contain letters, numbers, dot, underscore, and dash.');
        }

        return $site;
    }

    private function assertSnapshot(string $snapshot): void
    {
        if (preg_match('#^\d{4}/\d{2}/\d{2}/\d{6}-[0-9a-f-]{36}$#', $snapshot) !== 1) {
            throw new RuntimeException('Invalid backup snapshot identifier.');
        }
    }

    private function assertSafeRelativePath(string $path): void
    {
        if ($path === '' || str_starts_with($path, '/') || str_contains($path, '..') || str_contains($path, '\\')) {
            throw new RuntimeException("Unsafe path in backup manifest: {$path}");
        }
    }
}
