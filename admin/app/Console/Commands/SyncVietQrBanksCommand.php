<?php

namespace App\Console\Commands;

use App\Services\Admin\VietQrBankService;
use Illuminate\Console\Command;
use Illuminate\Support\Facades\Cache;
use Throwable;

class SyncVietQrBanksCommand extends Command
{
    protected $signature = 'banks:sync-vietqr';

    protected $description = 'Đồng bộ danh sách ngân hàng VietQR vào DB, cache và Redis chia sẻ';

    public function handle(VietQrBankService $service): int
    {
        $lock = Cache::store($service->cacheStore())->lock('lock:banks:sync-vietqr', 3600);

        if (! $lock->get()) {
            $this->warn('Đang có một tiến trình đồng bộ danh sách ngân hàng khác chạy.');

            return self::SUCCESS;
        }

        try {
            $snapshot = $service->syncFromProvider();

            $this->info(sprintf(
                'Đã đồng bộ %d ngân hàng từ VietQR. Redis key: %s',
                $snapshot['count'],
                $snapshot['redis_key'],
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
