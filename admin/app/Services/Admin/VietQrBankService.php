<?php

namespace App\Services\Admin;

use App\Models\Payment\VietQrBank;
use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Redis;
use RuntimeException;

class VietQrBankService
{
    public function getSnapshot(): array
    {
        $snapshot = Cache::store($this->cacheStore())->get($this->cacheKey());

        if (is_array($snapshot) && $this->runtimeRedisHasSnapshot()) {
            return $snapshot;
        }

        return $this->primeRuntimeStoresFromDatabase();
    }

    public function syncFromProvider(): array
    {
        $payload = $this->fetchPayload();
        $rows = $this->normalizeRows($payload);

        if ($rows === []) {
            throw new RuntimeException('Danh sách ngân hàng VietQR không hợp lệ.');
        }

        DB::transaction(function () use ($rows): void {
            VietQrBank::query()->upsert(
                $rows,
                ['code'],
                [
                    'source_id',
                    'name',
                    'short_name',
                    'bin',
                    'logo',
                    'transfer_supported',
                    'lookup_supported',
                    'support',
                    'raw_payload',
                    'synced_at',
                    'updated_at',
                ],
            );

            $codes = array_column($rows, 'code');

            if ($codes !== []) {
                VietQrBank::query()
                    ->whereNotIn('code', $codes)    
                    ->delete();
            }
        });

        return $this->primeRuntimeStoresFromDatabase();
    }

    public function primeRuntimeStoresFromDatabase(): array
    {
        $snapshot = $this->buildSnapshotFromDatabase();

        return $this->primeRuntimeStores($snapshot);
    }

    public function primeRuntimeStores(array $snapshot): array
    {
        Cache::store($this->cacheStore())->put(
            $this->cacheKey(),
            $snapshot,
            now()->addSeconds($this->cacheTtlSeconds()),
        );

        $redis = Redis::connection($this->redisConnection());

        $redis->set($this->redisKey(), json_encode($snapshot, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES));
        $redis->expire($this->redisKey(), $this->cacheTtlSeconds());

        return $snapshot;
    }

    public function forgetRuntimeStores(): void
    {
        Cache::store($this->cacheStore())->forget($this->cacheKey());
        Redis::connection($this->redisConnection())->del($this->redisKey());
    }

    public function cacheStore(): string
    {
        return (string) config('services.vietqr_banks.cache_store', 'redis');
    }

    public function cacheKey(): string
    {
        return (string) config('services.vietqr_banks.cache_key', 'admin:vietqr:banks:snapshot');
    }

    public function redisConnection(): string
    {
        return (string) config('services.vietqr_banks.redis_connection', 'shared');
    }

    public function redisKey(): string
    {
        return (string) config('services.vietqr_banks.redis_key', 'shared:vietqr:banks');
    }

    public function cacheTtlSeconds(): int
    {
        return (int) config('services.vietqr_banks.cache_ttl_seconds', 86400);
    }

    private function fetchPayload(): array
    {
        $response = Http::timeout((int) config('services.vietqr_banks.timeout', 10))
            ->acceptJson()
            ->retry((int) config('services.vietqr_banks.retry', 2), 250)
            ->get((string) config('services.vietqr_banks.url'));

        if (! $response->successful()) {
            throw new RuntimeException('Không thể tải danh sách ngân hàng từ VietQR.');
        }

        $payload = $response->json();

        if (! is_array($payload) || ($payload['code'] ?? null) !== '00') {
            throw new RuntimeException('VietQR trả về dữ liệu danh sách ngân hàng không hợp lệ.');
        }

        return $payload;
    }

    private function normalizeRows(array $payload): array
    {
        $rows = [];

        foreach ((array) data_get($payload, 'data', []) as $item) {
            if (! is_array($item)) {
                continue;
            }

            $sourceId = (int) data_get($item, 'id');
            $code = trim((string) data_get($item, 'code'));
            $bin = trim((string) data_get($item, 'bin'));

            if ($sourceId <= 0 || $code === '' || $bin === '') {
                continue;
            }

            $rows[] = [
                'source_id' => $sourceId,
                'code' => $code,
                'name' => trim((string) data_get($item, 'name', '')),
                'short_name' => trim((string) (data_get($item, 'shortName') ?? data_get($item, 'short_name') ?? '')),
                'bin' => $bin,
                'logo' => data_get($item, 'logo'),
                'transfer_supported' => (bool) data_get($item, 'transferSupported', data_get($item, 'isTransfer', false)),
                'lookup_supported' => (bool) data_get($item, 'lookupSupported', false),
                'support' => is_numeric(data_get($item, 'support')) ? (int) data_get($item, 'support') : null,
                'raw_payload' => json_encode($item, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES),
                'synced_at' => now(),
                'updated_at' => now(),
                'created_at' => now(),
            ];
        }

        return $rows;
    }

    private function buildSnapshotFromDatabase(): array
    {
        $banks = VietQrBank::query()
            ->orderBy('short_name')
            ->orderBy('name')
            ->get()
            ->map(static fn (VietQrBank $bank): array => [
                'source_id' => $bank->source_id,
                'code' => $bank->code,
                'name' => $bank->name,
                'short_name' => $bank->short_name,
                'bin' => $bank->bin,
                'logo' => $bank->logo,
                'transfer_supported' => (bool) $bank->transfer_supported,
                'lookup_supported' => (bool) $bank->lookup_supported,
                'support' => $bank->support,
                'synced_at' => $bank->synced_at?->toIso8601String(),
            ])
            ->values()
            ->all();

        $latestSync = VietQrBank::query()->latest('synced_at')->first()?->synced_at;

        return [
            'source_name' => config('services.vietqr_banks.source_name', 'vietqr'),
            'count' => count($banks),
            'synced_at' => $latestSync?->toIso8601String(),
            'cache_store' => $this->cacheStore(),
            'cache_key' => $this->cacheKey(),
            'redis_connection' => $this->redisConnection(),
            'redis_key' => $this->redisKey(),
            'banks' => $banks,
        ];
    }

    private function runtimeRedisHasSnapshot(): bool
    {
        return Redis::connection($this->redisConnection())->exists($this->redisKey()) > 0;
    }
}
