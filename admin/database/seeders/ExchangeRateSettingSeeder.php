<?php

namespace Database\Seeders;

use App\Models\System\ExchangeRateSetting;
use Illuminate\Database\Seeder;

class ExchangeRateSettingSeeder extends Seeder
{
    public function run(): void
    {
        ExchangeRateSetting::updateOrCreate(
            ['code' => ExchangeRateSetting::CODE],
            [
                'base_currency' => 'USDT',
                'quote_currency' => 'VND',
                'rate' => 25000,
                'source_rate' => 25000,
                'auto_sync' => true,
                'source_name' => 'seed',
                'note' => 'Giá khởi tạo để admin điều chỉnh sau khi triển khai.',
            ],
        );
    }
}
