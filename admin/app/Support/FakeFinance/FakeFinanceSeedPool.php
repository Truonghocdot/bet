<?php

namespace App\Support\FakeFinance;

use SimpleXMLElement;
use ZipArchive;

class FakeFinanceSeedPool
{
    /**
     * @return list<array{masked_code:string,masked_phone:string,status_label:string}>
     */
    public static function load(): array
    {
        $xlsxPath = base_path('../danh_sach.xlsx');

        if (is_file($xlsxPath)) {
            $rows = static::parseXlsxPool($xlsxPath);

            if ($rows !== []) {
                return $rows;
            }
        }

        return static::fallbackPool();
    }

    /**
     * @return list<array{masked_code:string,masked_phone:string,status_label:string}>
     */
    private static function parseXlsxPool(string $path): array
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

        $namespace = 'http://schemas.openxmlformats.org/spreadsheetml/2006/main';
        $sheetData = $worksheet->children($namespace)->sheetData;
        if (! $sheetData instanceof SimpleXMLElement) {
            return [];
        }

        $rows = $sheetData->children($namespace);

        $pool = [];

        foreach ($rows as $row) {
            if ($row->getName() !== 'row') {
                continue;
            }

            $rowNumber = (int) ((string) ($row['r'] ?? '0'));
            if ($rowNumber === 1) {
                continue;
            }

            $values = [];

            foreach ($row->children($namespace) as $cell) {
                if ($cell->getName() !== 'c') {
                    continue;
                }

                $inlineString = $cell->children($namespace)->is;
                $inlineText = $inlineString instanceof SimpleXMLElement
                    ? trim((string) $inlineString->children($namespace)->t)
                    : '';

                $value = $inlineText !== '' ? $inlineText : trim((string) $cell->children($namespace)->v);

                if ($value !== '') {
                    $values[] = $value;
                }
            }

            if (count($values) < 3) {
                continue;
            }

            // Defensive guard: skip the header row even if row-number parsing behaves differently across environments.
            if (
                mb_strtolower($values[0]) === 'id'
                && mb_strtolower($values[1]) === 'số điện thoại'
                && mb_strtolower($values[2]) === 'trạng thái'
            ) {
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
    private static function fallbackPool(): array
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
}
