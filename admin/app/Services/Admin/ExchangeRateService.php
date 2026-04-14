<?php

namespace App\Services\Admin;

use App\Models\System\ExchangeRateSetting;
use App\Models\User;
use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Redis;
use Illuminate\Support\Str;
use RuntimeException;

class ExchangeRateService
{
    public function getSnapshot(): array
    {
        $snapshot = Cache::store($this->cacheStore())->get($this->cacheKey());

        if (is_array($snapshot) && $this->runtimeRedisHasSnapshot()) {
            return $snapshot;
        }

        return $this->primeRuntimeStores($this->setting());
    }

    public function setting(): ExchangeRateSetting
    {
        return ExchangeRateSetting::query()->firstOrCreate(
            ['code' => ExchangeRateSetting::CODE],
            [
                'base_currency' => 'USDT',
                'quote_currency' => 'VND',
                'rate' => 25000,
                'source_rate' => 25000,
                'auto_sync' => true,
                'source_name' => 'seed',
            ],
        );
    }

    public function saveSetting(array $data, ?User $actor = null): ExchangeRateSetting
    {
        $setting = DB::transaction(function () use ($data, $actor): ExchangeRateSetting {
            $setting = $this->setting();

            $setting->fill([
                'rate' => $data['rate'],
                'source_rate' => $data['source_rate'] ?? $data['rate'],
                'auto_sync' => (bool) ($data['auto_sync'] ?? false),
                'source_name' => $data['source_name'] ?? 'manual',
                'last_synced_at' => now(),
                'note' => $data['note'] ?? null,
                'nowpayments_api_key' => $data['nowpayments_api_key'] ?? null,
                'nowpayments_ipn_secret' => $data['nowpayments_ipn_secret'] ?? null,
                'nowpayments_payout_wallet' => $data['nowpayments_payout_wallet'] ?? null,
                'nowpayments_sandbox' => (bool) ($data['nowpayments_sandbox'] ?? false),
                'telegram_cskh_link' => $data['telegram_cskh_link'] ?? null,
                'updated_by' => $actor?->id,
            ]);

            $setting->save();

            return $setting->fresh();
        });

        $this->primeRuntimeStores($setting);

        return $setting->fresh();
    }

    public function refreshFromProvider(): ExchangeRateSetting
    {
        $setting = $this->setting();
        $providerUrl = config('services.exchange_rate.usdt_vnd_url');
        $sourceName = config('services.exchange_rate.source_name', 'provider');

        if (blank($providerUrl)) {
            $setting = DB::transaction(function () use ($setting): ExchangeRateSetting {
                $setting->forceFill([
                    'source_rate' => $setting->rate,
                    'source_name' => 'manual',
                    'last_synced_at' => now(),
                ])->save();

                return $setting->fresh();
            });

            $this->primeRuntimeStores($setting);

            return $setting->fresh();
        }

        $response = Http::timeout((int) config('services.exchange_rate.timeout', 10))
            ->acceptJson()
            ->retry((int) config('services.exchange_rate.retry', 2), 250)
            ->get($providerUrl);

        if (! $response->successful()) {
            throw new RuntimeException('Không thể đồng bộ tỉ giá USDT/VND từ nguồn cung cấp.');
        }

        $rate = $this->extractRateFromPayload($response->json() ?? $response->body());

        if ($rate <= 0) {
            throw new RuntimeException('Nguồn cung cấp không trả về tỉ giá hợp lệ.');
        }

        $setting = DB::transaction(function () use ($setting, $rate, $sourceName): ExchangeRateSetting {
            $setting->forceFill([
                'source_rate' => $rate,
                'source_name' => $sourceName,
                'last_synced_at' => now(),
                'rate' => $setting->auto_sync ? $rate : $setting->rate,
            ])->save();

            return $setting->fresh();
        });

        $this->primeRuntimeStores($setting);

        return $setting->fresh();
    }

    public function primeRuntimeStores(ExchangeRateSetting $setting): array
    {
        $payload = $this->buildSnapshot($setting);

        Cache::store($this->cacheStore())->put(
            $this->cacheKey(),
            $payload,
            now()->addSeconds($this->cacheTtlSeconds()),
        );

        $redis = Redis::connection($this->redisConnection());

        $redis->set($this->redisKey(), json_encode($payload, JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES));
        $redis->expire($this->redisKey(), $this->cacheTtlSeconds());

        return $payload;
    }

    public function forgetRuntimeStores(): void
    {
        Cache::store($this->cacheStore())->forget($this->cacheKey());
        Redis::connection($this->redisConnection())->del($this->redisKey());
    }

    public function syncRuntimeStoresFromDatabase(): ExchangeRateSetting
    {
        $setting = $this->setting();
        $this->primeRuntimeStores($setting);

        return $setting;
    }

    public function cacheTtlSeconds(): int
    {
        return (int) config('services.exchange_rate.cache_ttl_seconds', 300);
    }

    public function cacheStore(): string
    {
        return (string) config('services.exchange_rate.cache_store', 'redis');
    }

    public function cacheKey(): string
    {
        return (string) config('services.exchange_rate.cache_key', 'admin:exchange-rate:usdt-vnd:snapshot');
    }

    public function redisConnection(): string
    {
        return (string) config('services.exchange_rate.redis_connection', 'shared');
    }

    public function redisKey(): string
    {
        return (string) config('services.exchange_rate.redis_key', 'shared:exchange-rate:usdt-vnd');
    }

    private function buildSnapshot(ExchangeRateSetting $setting): array
    {
        return [
            'code' => $setting->code,
            'base_currency' => $setting->base_currency,
            'quote_currency' => $setting->quote_currency,
            'rate' => (string) $setting->rate,
            'source_rate' => $setting->source_rate !== null ? (string) $setting->source_rate : null,
            'auto_sync' => (bool) $setting->auto_sync,
            'source_name' => $setting->source_name,
            'last_synced_at' => $setting->last_synced_at?->toIso8601String(),
            'updated_at' => $setting->updated_at?->toIso8601String(),
            'note' => $setting->note,
            'nowpayments_api_key' => $setting->nowpayments_api_key,
            'nowpayments_ipn_secret' => $setting->nowpayments_ipn_secret,
            'nowpayments_payout_wallet' => $setting->nowpayments_payout_wallet,
            'nowpayments_sandbox' => (bool) $setting->nowpayments_sandbox,
            'telegram_cskh_link' => $setting->telegram_cskh_link,
            'cache_store' => $this->cacheStore(),
            'cache_key' => $this->cacheKey(),
            'redis_connection' => $this->redisConnection(),
            'redis_key' => $this->redisKey(),
        ];
    }

    private function runtimeRedisHasSnapshot(): bool
    {
        return Redis::connection($this->redisConnection())->exists($this->redisKey()) > 0;
    }

    private function extractRateFromPayload(array|string $payload): float
    {
        if (is_string($payload) && is_numeric(trim($payload))) {
            return (float) trim($payload);
        }

        if (! is_array($payload)) {
            return 0.0;
        }

        $candidates = [
            data_get($payload, 'tether.vnd'),
            data_get($payload, 'rate'),
            data_get($payload, 'price'),
            data_get($payload, 'data.rate'),
            data_get($payload, 'data.price'),
            data_get($payload, 'result.rate'),
            data_get($payload, 'result.price'),
            data_get($payload, 'data[0].rate'),
            data_get($payload, 'data[0].price'),
        ];

        foreach ($candidates as $candidate) {
            if (is_numeric($candidate)) {
                return (float) $candidate;
            }

            if (is_string($candidate)) {
                $normalized = (float) Str::of($candidate)->replace(',', '');

                if ($normalized > 0) {
                    return $normalized;
                }
            }
        }

        return 0.0;
    }
}
