<?php

namespace App\Console\Commands;

use App\Models\Finance\FakeDepositTransaction;
use App\Models\Finance\FakeWithdrawTransaction;
use Illuminate\Console\Command;
use Illuminate\Support\Facades\DB;
use SimpleXMLElement;
use ZipArchive;

class GenerateFakeFinanceTransactionsCommand extends Command
{
    protected $signature = 'finance:generate-fake-transactions';

    protected $description = 'Generate fake deposit and withdraw transaction feeds from the provided sample pool';

    private const MAX_ROWS_PER_TABLE = 300;

    public function handle(): int
    {
        $pool = $this->loadSeedPool();

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
     * @return list<array{masked_code:string,masked_phone:string,status_label:string}>
     */
    private function loadSeedPool(): array
    {
        $xlsxPath = base_path('../danh_sach.xlsx');

        if (is_file($xlsxPath)) {
            $rows = $this->parseXlsxPool($xlsxPath);

            if ($rows !== []) {
                return $rows;
            }
        }

        return $this->fallbackPool();
    }

    /**
     * @return list<array{masked_code:string,masked_phone:string,status_label:string}>
     */
    private function parseXlsxPool(string $path): array
    {
        if (! class_exists(ZipArchive::class)) {
            return [];
        }

        $archive = new ZipArchive();
        if ($archive->open($path) !== true) {
            return [];
        }

        $worksheetXml = $archive->getFromName('xl/worksheets/sheet1.xml') ?: '';
        $archive->close();

        if ($worksheetXml === '') {
            return [];
        }

        $worksheet = simplexml_load_string($worksheetXml);
        if (! $worksheet instanceof SimpleXMLElement) {
            return [];
        }

        $worksheet->registerXPathNamespace('sheet', 'http://schemas.openxmlformats.org/spreadsheetml/2006/main');
        $rows = $worksheet->xpath('//sheet:sheetData/sheet:row') ?: [];

        $pool = [];

        foreach ($rows as $row) {
            if ((int) ($row['r'] ?? 0) === 1) {
                continue;
            }

            $cells = $row->xpath('./sheet:c') ?: [];
            $values = [];

            foreach ($cells as $cell) {
                $inlineText = $cell->xpath('./sheet:is/sheet:t');
                $value = $inlineText !== [] ? trim((string) $inlineText[0]) : trim((string) $cell->v);

                if ($value !== '') {
                    $values[] = $value;
                }
            }

            if (count($values) < 3) {
                continue;
            }

            $pool[] = [
                'masked_code' => $values[0],
                'masked_phone' => $values[1],
                'status_label' => $values[2],
            ];
        }

        return $pool;
    }

    /**
     * @return list<array{masked_code:string,masked_phone:string,status_label:string}>
     */
    private function fallbackPool(): array
    {
        return [
            ['masked_code' => '**0541', 'masked_phone' => '+84*****8480', 'status_label' => 'Thành công'],
            ['masked_code' => '**0942', 'masked_phone' => '+84*****2978', 'status_label' => 'Thành công'],
            ['masked_code' => '**2574', 'masked_phone' => '+84*****5221', 'status_label' => 'Thành công'],
            ['masked_code' => '**1990', 'masked_phone' => '+84*****8825', 'status_label' => 'Thành công'],
            ['masked_code' => '**7458', 'masked_phone' => '+84*****5774', 'status_label' => 'Thành công'],
            ['masked_code' => '**1111', 'masked_phone' => '+84*****4928', 'status_label' => 'Thành công'],
            ['masked_code' => '**8890', 'masked_phone' => '+84*****6799', 'status_label' => 'Thành công'],
            ['masked_code' => '**1237', 'masked_phone' => '+84*****5015', 'status_label' => 'Thành công'],
            ['masked_code' => '**6192', 'masked_phone' => '+84*****1047', 'status_label' => 'Thành công'],
            ['masked_code' => '**4774', 'masked_phone' => '+84*****5924', 'status_label' => 'Thành công'],
            ['masked_code' => '**0906', 'masked_phone' => '+84*****1763', 'status_label' => 'Thành công'],
            ['masked_code' => '**5670', 'masked_phone' => '+84*****3732', 'status_label' => 'Thành công'],
            ['masked_code' => '**2577', 'masked_phone' => '+84*****9768', 'status_label' => 'Thành công'],
            ['masked_code' => '**4586', 'masked_phone' => '+84*****4652', 'status_label' => 'Thành công'],
            ['masked_code' => '**4383', 'masked_phone' => '+84*****8267', 'status_label' => 'Thành công'],
            ['masked_code' => '**6772', 'masked_phone' => '+84*****0406', 'status_label' => 'Thành công'],
            ['masked_code' => '**2804', 'masked_phone' => '+84*****9551', 'status_label' => 'Thành công'],
            ['masked_code' => '**2896', 'masked_phone' => '+84*****1789', 'status_label' => 'Thành công'],
            ['masked_code' => '**9081', 'masked_phone' => '+84*****6076', 'status_label' => 'Thành công'],
        ];
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
