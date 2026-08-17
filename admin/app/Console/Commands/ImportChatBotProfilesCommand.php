<?php

namespace App\Console\Commands;

use App\Models\Chat\ChatBotProfile;
use Illuminate\Console\Command;

class ImportChatBotProfilesCommand extends Command
{
    protected $signature = 'chat:import-bot-profiles {file : Đường dẫn CSV hoặc JSON} {--dry-run}';

    protected $description = 'Import hoặc cập nhật bot profile chat global';

    public function handle(): int
    {
        $path = (string) $this->argument('file');
        if (! is_file($path) || ! is_readable($path)) {
            $this->error('Không đọc được file: '.$path);

            return self::FAILURE;
        }

        $rows = strtolower(pathinfo($path, PATHINFO_EXTENSION)) === 'json'
            ? $this->readJson($path)
            : $this->readCsv($path);
        if ($rows === null) {
            return self::FAILURE;
        }

        $created = $updated = $invalid = 0;
        foreach ($rows as $row) {
            $displayName = trim((string) ($row['display_name'] ?? ''));
            if ($displayName === '' || mb_strlen($displayName) > 120) {
                $invalid++;

                continue;
            }

            $active = filter_var($row['active'] ?? true, FILTER_VALIDATE_BOOL, FILTER_NULL_ON_FAILURE) ?? true;
            $payload = [
                'avatar_path' => filled($row['avatar_path'] ?? null) ? trim((string) $row['avatar_path']) : null,
                'active' => $active,
                'sort_order' => max(0, (int) ($row['sort_order'] ?? 0)),
            ];
            $existing = ChatBotProfile::query()->where('display_name', $displayName)->first();
            if ($existing) {
                $updated++;
                if (! $this->option('dry-run')) {
                    $existing->fill($payload)->save();
                }

                continue;
            }

            $created++;
            if (! $this->option('dry-run')) {
                ChatBotProfile::query()->create(['display_name' => $displayName, ...$payload]);
            }
        }

        $mode = $this->option('dry-run') ? 'Dry run' : 'Import';
        $this->info("{$mode}: {$created} thêm, {$updated} cập nhật, {$invalid} không hợp lệ.");

        return $invalid > 0 ? self::FAILURE : self::SUCCESS;
    }

    /** @return array<int, array<string, mixed>>|null */
    private function readJson(string $path): ?array
    {
        $decoded = json_decode((string) file_get_contents($path), true);
        if (! is_array($decoded)) {
            $this->error('JSON phải là một mảng các bot profile.');

            return null;
        }

        return array_values(array_filter($decoded, 'is_array'));
    }

    /** @return array<int, array<string, mixed>>|null */
    private function readCsv(string $path): ?array
    {
        $handle = fopen($path, 'rb');
        if (! $handle) {
            $this->error('Không thể mở CSV.');

            return null;
        }

        $headers = fgetcsv($handle);
        if (! is_array($headers)) {
            fclose($handle);
            $this->error('CSV không có dòng tiêu đề.');

            return null;
        }

        $headers = array_map(fn ($header): string => strtolower(trim((string) $header)), $headers);
        $rows = [];
        while (($values = fgetcsv($handle)) !== false) {
            $rows[] = array_combine($headers, array_pad(array_slice($values, 0, count($headers)), count($headers), null)) ?: [];
        }
        fclose($handle);

        return $rows;
    }
}
