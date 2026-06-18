<?php

namespace App\Console\Commands;

use App\Models\Finance\FakeDepositTransaction;
use App\Models\Finance\FakeWithdrawTransaction;
use App\Support\FakeFinance\FakeFinanceSeedPool;
use Illuminate\Console\Command;
use Illuminate\Support\Facades\DB;

class GenerateFakeFinanceTransactionsCommand extends Command
{
    protected $signature = 'finance:generate-fake-transactions';

    protected $description = 'Generate fake deposit and withdraw transaction feeds from the provided sample pool';

    private const MAX_ROWS_PER_TABLE = 300;

    public function handle(): int
    {
        $pool = FakeFinanceSeedPool::load();

        if ($pool === []) {
            $this->warn('Khong tim thay pool fake data de sinh giao dich.');

            return self::SUCCESS;
        }

        $depositInserted = $this->appendBatch(FakeDepositTransaction::class, $pool, 'deposit', random_int(1, 3), 0);
        $withdrawInserted = $this->appendBatch(FakeWithdrawTransaction::class, $pool, 'withdraw', random_int(1, 3), intdiv(count($pool), 2));

        $this->info(sprintf(
            'Da them %d giao dich nap fake va %d giao dich rut fake.',
            $depositInserted,
            $withdrawInserted,
        ));

        return self::SUCCESS;
    }
    /**
     * @param  class-string<FakeDepositTransaction|FakeWithdrawTransaction>  $modelClass
     * @param  list<array{masked_code:string,masked_phone:string,status_label:string}>  $pool
     */
    private function appendBatch(string $modelClass, array $pool, string $channel, int $batchSize, int $seedOffset): int
    {
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
                'meta' => json_encode([
                    'channel' => $channel,
                    'source_file' => 'danh_sach.xlsx',
                    'source_index' => $sourceIndex,
                ], JSON_UNESCAPED_UNICODE | JSON_UNESCAPED_SLASHES),
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
