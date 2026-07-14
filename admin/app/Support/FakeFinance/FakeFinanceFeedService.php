<?php

namespace App\Support\FakeFinance;

use App\Models\Finance\FakeDepositTransaction;
use App\Models\Finance\FakeWithdrawTransaction;
use Illuminate\Support\Facades\DB;

class FakeFinanceFeedService
{
    private const MAX_ROWS_PER_TABLE = 300;

    public function appendDepositBatch(int $batchSize = 1, array $meta = []): int
    {
        return $this->appendBatch(
            FakeDepositTransaction::class,
            'deposit',
            $batchSize,
            0,
            $meta,
        );
    }

    public function appendWithdrawBatch(int $batchSize = 1, array $meta = []): int
    {
        $pool = FakeFinanceSeedPool::load();

        return $this->appendBatch(
            FakeWithdrawTransaction::class,
            'withdraw',
            $batchSize,
            intdiv(count($pool), 2),
            $meta,
            $pool,
        );
    }

    /**
     * @param  class-string<FakeDepositTransaction|FakeWithdrawTransaction>  $modelClass
     * @param  list<array{masked_code:string,masked_phone:string,status_label:string}>|null  $pool
     */
    private function appendBatch(
        string $modelClass,
        string $channel,
        int $batchSize,
        int $seedOffset,
        array $meta = [],
        ?array $pool = null,
    ): int {
        $batchSize = max(0, $batchSize);
        if ($batchSize === 0) {
            return 0;
        }

        $pool ??= FakeFinanceSeedPool::load();
        $poolCount = count($pool);

        if ($poolCount === 0) {
            return 0;
        }

        $lastSourceIndex = (int) ($modelClass::query()->latest('id')->value('source_index') ?? 0);
        $startIndex = $lastSourceIndex > 0
            ? (($lastSourceIndex % $poolCount) + 1)
            : (($seedOffset % $poolCount) + 1);

        $now = now();
        $rows = [];

        for ($offset = 0; $offset < $batchSize; $offset++) {
            $sourceIndex = (($startIndex - 1 + $offset) % $poolCount) + 1;
            $entry = $pool[$sourceIndex - 1];

            $rows[] = [
                'source_index' => $sourceIndex,
                'masked_code' => $entry['masked_code'],
                'masked_phone' => $entry['masked_phone'],
                'status_label' => $entry['status_label'],
                'meta' => json_encode(array_merge([
                    'channel' => $channel,
                    'source_file' => 'danh_sach.xlsx',
                    'source_index' => $sourceIndex,
                ], $meta), JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES),
                'created_at' => $now,
                'updated_at' => $now,
            ];
        }

        DB::transaction(function () use ($modelClass, $rows): void {
            $modelClass::query()->insert($rows);
            $this->trimOverflow($modelClass);
        });

        return count($rows);
    }

    /**
     * @param  class-string<FakeDepositTransaction|FakeWithdrawTransaction>  $modelClass
     */
    private function trimOverflow(string $modelClass): void
    {
        $overflow = $modelClass::query()->count() - self::MAX_ROWS_PER_TABLE;

        if ($overflow <= 0) {
            return;
        }

        $ids = $modelClass::query()
            ->orderBy('id')
            ->limit($overflow)
            ->pluck('id');

        if ($ids->isEmpty()) {
            return;
        }

        $modelClass::query()->whereKey($ids)->delete();
    }
}
