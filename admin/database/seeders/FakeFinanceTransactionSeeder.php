<?php

namespace Database\Seeders;

use App\Models\Finance\FakeDepositTransaction;
use App\Models\Finance\FakeWithdrawTransaction;
use App\Support\FakeFinance\FakeFinanceSeedPool;
use Illuminate\Database\Seeder;

class FakeFinanceTransactionSeeder extends Seeder
{
    private const TARGET_RECORDS = 10000;
    private const CHUNK_SIZE = 1000;

    public function run(): void
    {
        $pool = FakeFinanceSeedPool::load();

        if ($pool === []) {
            $this->command?->warn('Bo qua FakeFinanceTransactionSeeder vi khong co pool du lieu fake.');

            return;
        }

        FakeDepositTransaction::query()->truncate();
        FakeWithdrawTransaction::query()->truncate();

        $this->seedTable(FakeDepositTransaction::class, $pool, 'deposit', 0);
        $this->seedTable(FakeWithdrawTransaction::class, $pool, 'withdraw', intdiv(count($pool), 2));

        $this->command?->info('Da seed 10.000 giao dich nap fake va 10.000 giao dich rut fake.');
    }

    /**
     * @param  class-string<FakeDepositTransaction|FakeWithdrawTransaction>  $modelClass
     * @param  list<array{masked_code:string,masked_phone:string,status_label:string}>  $pool
     */
    private function seedTable(string $modelClass, array $pool, string $channel, int $seedOffset): void
    {
        $poolCount = count($pool);

        for ($batchStart = 0; $batchStart < self::TARGET_RECORDS; $batchStart += self::CHUNK_SIZE) {
            $rows = [];
            $limit = min(self::CHUNK_SIZE, self::TARGET_RECORDS - $batchStart);

            for ($offset = 0; $offset < $limit; $offset++) {
                $rowIndex = $batchStart + $offset;
                $sourceIndex = (($seedOffset + $rowIndex) % $poolCount) + 1;
                $entry = $pool[$sourceIndex - 1];
                $createdAt = now()->subSeconds(self::TARGET_RECORDS - $rowIndex);

                $rows[] = [
                    'source_index' => $sourceIndex,
                    'masked_code' => $entry['masked_code'],
                    'masked_phone' => $entry['masked_phone'],
                    'status_label' => $entry['status_label'],
                    'meta' => [
                        'channel' => $channel,
                        'source_file' => 'danh_sach.xlsx',
                        'source_index' => $sourceIndex,
                        'seed_batch' => intdiv($batchStart, self::CHUNK_SIZE) + 1,
                    ],
                    'created_at' => $createdAt,
                    'updated_at' => $createdAt,
                ];
            }

            $modelClass::query()->insert($rows);
        }
    }
}
