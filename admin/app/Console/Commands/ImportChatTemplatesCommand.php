<?php

namespace App\Console\Commands;

use App\Models\Chat\ChatBotProfile;
use App\Models\Chat\ChatBotTemplate;
use Illuminate\Console\Command;

class ImportChatTemplatesCommand extends Command
{
    protected $signature = 'chat:import-templates {file : Đường dẫn CSV hoặc JSON} {--dry-run} {--bot-profile-id=}';

    protected $description = 'Import kho câu mẫu chat global';

    public function handle(): int
    {
        $path = (string) $this->argument('file');
        if (! is_file($path) || ! is_readable($path)) {
            $this->error('Không đọc được file: '.$path);

            return self::FAILURE;
        }

        $botProfileID = $this->option('bot-profile-id');
        if ($botProfileID && ! ChatBotProfile::query()->whereKey($botProfileID)->exists()) {
            $this->error('Không tìm thấy bot profile ID '.$botProfileID.'.');

            return self::FAILURE;
        }

        $extension = strtolower(pathinfo($path, PATHINFO_EXTENSION));
        $rows = $extension === 'json' ? $this->readJson($path) : $this->readCsv($path);
        if ($rows === null) {
            return self::FAILURE;
        }

        $created = $skipped = $invalid = 0;
        foreach ($rows as $row) {
            $body = trim((string) ($row['body'] ?? ''));
            $category = trim((string) ($row['category'] ?? 'general')) ?: 'general';
            $language = trim((string) ($row['language'] ?? 'vi')) ?: 'vi';
            $active = filter_var($row['active'] ?? true, FILTER_VALIDATE_BOOL, FILTER_NULL_ON_FAILURE);
            $active = $active ?? true;
            if ($body === '' || mb_strlen($body) > 280 || preg_match('/(https?:\/\/|www\.|(?:^|[\s(])(?:[a-z0-9-]+\.)+[a-z]{2,}(?:[\/?#:)]|\s|$))/iu', $body)) {
                $invalid++;

                continue;
            }

            $query = ChatBotTemplate::query()
                ->where('body', $body)
                ->where('category', $category)
                ->where('language', $language);
            if ($query->exists()) {
                $skipped++;

                continue;
            }
            if (! $this->option('dry-run')) {
                ChatBotTemplate::query()->create([
                    'bot_profile_id' => $botProfileID ?: null,
                    'body' => $body,
                    'category' => $category,
                    'language' => $language,
                    'active' => $active,
                ]);
            }
            $created++;
        }

        $this->info(sprintf('%s: %d thêm, %d bỏ qua, %d không hợp lệ.', $this->option('dry-run') ? 'Dry run' : 'Import', $created, $skipped, $invalid));

        return $invalid > 0 ? self::FAILURE : self::SUCCESS;
    }

    /** @return array<int, array<string, mixed>>|null */
    private function readJson(string $path): ?array
    {
        $decoded = json_decode((string) file_get_contents($path), true);
        if (! is_array($decoded)) {
            $this->error('JSON phải là một mảng các câu mẫu.');

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
