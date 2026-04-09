<?php

namespace App\Console\Commands;

use App\Services\Admin\ExchangeRateService;
use Illuminate\Console\Command;
use Illuminate\Support\Facades\Cache;
use Throwable;

class SyncUsdtVndRateCommand extends Command
{
    protected $signature = 'rates:sync-usdt-vnd';

    protected $description = 'Đồng bộ tỉ giá USDT/VND từ nguồn và đẩy vào cache/redis';

    public function handle(ExchangeRateService $service): int
    {
        $lock = Cache::store('redis')->lock('lock:rates:sync-usdt-vnd', 240);

        if (! $lock->get()) {
            $this->warn('Đang có một tiến trình đồng bộ tỉ giá khác chạy.');
            return self::SUCCESS;
        }

        try {
            $setting = $service->refreshFromProvider();

            $this->info(sprintf(
                'Đã đồng bộ USDT/VND. Tỷ giá áp dụng: %s VND',
                number_format((float) $setting->rate, 0, ',', '.'),
            ));

            return self::SUCCESS;
        } catch (Throwable $exception) {
            $this->error($exception->getMessage());

            return self::FAILURE;
        } finally {
            optional($lock)->release();
        }
    }
}
