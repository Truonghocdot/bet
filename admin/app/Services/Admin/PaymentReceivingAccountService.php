<?php

namespace App\Services\Admin;

use App\Enum\Payment\PaymentReceivingAccountStatus;
use App\Models\Payment\PaymentReceivingAccount;
use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\Redis;

class PaymentReceivingAccountService
{
    public function getSnapshot(): array
    {
        $snapshot = Cache::store($this->cacheStore())->get($this->cacheKey());

        if (is_array($snapshot) && $this->runtimeRedisHasSnapshot()) {
            return $snapshot;
        }

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
        return (string) config('services.payment_receiving_accounts.cache_store', 'redis');
    }

    public function cacheKey(): string
    {
        return (string) config('services.payment_receiving_accounts.cache_key', 'admin:payment:receiving-accounts:snapshot');
    }

    public function cacheTtlSeconds(): int
    {
        return (int) config('services.payment_receiving_accounts.cache_ttl_seconds', 300);
    }

    public function redisConnection(): string
    {
        return (string) config('services.payment_receiving_accounts.redis_connection', 'shared');
    }

    public function redisKey(): string
    {
        return (string) config('services.payment_receiving_accounts.redis_key', 'shared:payment:receiving-accounts:v1');
    }

    private function buildSnapshotFromDatabase(): array
    {
        $accounts = PaymentReceivingAccount::query()
            ->where('status', PaymentReceivingAccountStatus::ACTIVE->value)
            ->orderByDesc('is_default')
            ->orderBy('unit')
            ->orderBy('type')
            ->orderBy('sort_order')
            ->orderBy('id')
            ->get()
            ->map(static function (PaymentReceivingAccount $account): array {
                return [
                    'id' => $account->id,
                    'code' => $account->code,
                    'name' => $account->name,
                    'type' => $account->type?->value ?? $account->type,
                    'unit' => $account->unit?->value ?? $account->unit,
                    'provider_code' => $account->provider_code,
                    'account_name' => $account->account_name,
                    'account_number' => $account->account_number,
                    'wallet_address' => $account->wallet_address,
                    'network' => $account->network,
                    'qr_code_path' => $account->qr_code_path,
                    'instructions' => $account->instructions,
                    'status' => $account->status?->value ?? $account->status,
                    'is_default' => (bool) $account->is_default,
                    'sort_order' => (int) $account->sort_order,
                ];
            })
            ->values()
            ->all();

        return [
            'source_name' => config('services.payment_receiving_accounts.source_name', 'payment_receiving_accounts'),
            'count' => count($accounts),
            'synced_at' => now()->toIso8601String(),
            'cache_store' => $this->cacheStore(),
            'cache_key' => $this->cacheKey(),
            'redis_connection' => $this->redisConnection(),
            'redis_key' => $this->redisKey(),
            'accounts' => $accounts,
        ];
    }

    private function runtimeRedisHasSnapshot(): bool
    {
        return Redis::connection($this->redisConnection())->exists($this->redisKey()) > 0;
    }
}
