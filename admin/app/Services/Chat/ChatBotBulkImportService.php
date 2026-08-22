<?php

namespace App\Services\Chat;

use App\Models\Chat\ChatBotProfile;
use App\Models\Chat\ChatBotTemplate;
use Illuminate\Support\Facades\DB;

class ChatBotBulkImportService
{
    /**
     * @return array{lines:int,profiles_created:int,profiles_updated:int,templates_created:int,duplicates:int,invalid:int,profiles_deactivated:int}
     */
    public function import(
        string $input,
        string $category = 'event',
        string $language = 'vi',
        bool $active = true,
        bool $deactivateLegacyProfiles = true,
    ): array {
        $category = mb_substr(trim($category) ?: 'event', 0, 60);
        $language = mb_substr(trim($language) ?: 'vi', 0, 12);
        $result = [
            'lines' => 0,
            'profiles_created' => 0,
            'profiles_updated' => 0,
            'templates_created' => 0,
            'duplicates' => 0,
            'invalid' => 0,
            'profiles_deactivated' => 0,
        ];

        $lines = preg_split('/\R/u', $input) ?: [];

        return DB::transaction(function () use ($lines, $category, $language, $active, $deactivateLegacyProfiles, $result): array {
            $profilesByGameID = $this->profilesByGameID();
            $importedGameIDs = [];

            foreach ($lines as $rawLine) {
                $line = trim((string) $rawLine);
                if ($line === '') {
                    continue;
                }

                $result['lines']++;
                $parsed = $this->parseLine($line);
                if ($parsed === null || ! $this->validBody($parsed['body'])) {
                    $result['invalid']++;

                    continue;
                }

                $profile = null;
                if ($parsed['game_id'] !== null) {
                    $gameID = $parsed['game_id'];
                    $importedGameIDs[$gameID] = true;
                    $profile = $profilesByGameID[$gameID] ?? null;
                    $displayName = $this->displayName($gameID);

                    if (! $profile) {
                        $profile = ChatBotProfile::query()->create([
                            'display_name' => $displayName,
                            'active' => true,
                            'sort_order' => count($profilesByGameID) + 1,
                        ]);
                        $profilesByGameID[$gameID] = $profile;
                        $result['profiles_created']++;
                    } elseif ($profile->display_name !== $displayName || ! $profile->active) {
                        $profile->forceFill(['display_name' => $displayName, 'active' => true])->save();
                        $result['profiles_updated']++;
                    }
                }

                $duplicateQuery = ChatBotTemplate::query()
                    ->where('body', $parsed['body'])
                    ->where('category', $category)
                    ->where('language', $language);
                $profile
                    ? $duplicateQuery->where('bot_profile_id', $profile->id)
                    : $duplicateQuery->whereNull('bot_profile_id');

                if ($duplicateQuery->exists()) {
                    $result['duplicates']++;

                    continue;
                }

                ChatBotTemplate::query()->create([
                    'bot_profile_id' => $profile?->id,
                    'body' => $parsed['body'],
                    'category' => $category,
                    'language' => $language,
                    'active' => $active,
                ]);
                $result['templates_created']++;
            }

            if ($deactivateLegacyProfiles && $importedGameIDs !== []) {
                $seenGameIDs = [];
                ChatBotProfile::query()
                    ->where('active', true)
                    ->orderBy('id')
                    ->get()
                    ->each(function (ChatBotProfile $profile) use (&$result, &$seenGameIDs): void {
                        $gameID = $this->gameID($profile->display_name);
                        $standardized = $gameID !== null && preg_match('/^ID game #[0-9]+$/', $profile->display_name) === 1;
                        if ($standardized && ! isset($seenGameIDs[$gameID])) {
                            $seenGameIDs[$gameID] = true;

                            return;
                        }

                        $profile->forceFill(['active' => false])->save();
                        $result['profiles_deactivated']++;
                    });
            }

            return $result;
        });
    }

    /** @return array{game_id:?string,body:string}|null */
    private function parseLine(string $line): ?array
    {
        $columns = preg_split('/\t+/u', $line, 2);
        if (is_array($columns) && count($columns) === 2) {
            $gameID = $this->gameID((string) $columns[0]);
            if ($gameID === null) {
                return null;
            }

            return ['game_id' => $gameID, 'body' => $this->normalizeBody((string) $columns[1])];
        }

        if (preg_match('/^(.+?#\s*[0-9]+)\s{2,}(.+)$/u', $line, $matches) === 1) {
            $gameID = $this->gameID($matches[1]);
            if ($gameID === null) {
                return null;
            }

            return ['game_id' => $gameID, 'body' => $this->normalizeBody($matches[2])];
        }

        return ['game_id' => null, 'body' => $this->normalizeBody($line)];
    }

    private function normalizeBody(string $body): string
    {
        return preg_replace('/\s+/u', ' ', trim($body)) ?? '';
    }

    private function validBody(string $body): bool
    {
        if ($body === '' || mb_strlen($body) > 280 || preg_match('/[<>]/u', $body) === 1) {
            return false;
        }

        return preg_match('/(https?:\/\/|www\.|(?:^|[\s(])(?:[a-z0-9-]+\.)+[a-z]{2,}(?:[\/?#:)]|\s|$))/iu', $body) !== 1;
    }

    /** @return array<string, ChatBotProfile> */
    private function profilesByGameID(): array
    {
        $profiles = [];

        foreach (ChatBotProfile::query()->orderByDesc('active')->orderBy('id')->get() as $profile) {
            $gameID = $this->gameID($profile->display_name);
            if ($gameID === null || isset($profiles[$gameID])) {
                continue;
            }

            $profiles[$gameID] = $profile;
        }

        return $profiles;
    }

    private function gameID(string $value): ?string
    {
        if (preg_match('/(?:ID\s*game|người\s*chơi)\s*#?\s*([0-9]{3,12})/iu', trim($value), $matches) !== 1) {
            return null;
        }

        return ltrim($matches[1], '0') ?: '0';
    }

    private function displayName(string $gameID): string
    {
        return 'ID game #'.$gameID;
    }
}
