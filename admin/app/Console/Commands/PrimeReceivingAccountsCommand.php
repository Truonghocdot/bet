<?php

namespace App\Console\Commands;

use App\Services\Admin\PaymentReceivingAccountService;
use Illuminate\Console\Command;

class PrimeReceivingAccountsCommand extends Command
{
    protected $signature = 'payment:prime-receiving-accounts';

    protected $description = 'Prime cache and shared Redis for payment receiving accounts';

    public function handle(PaymentReceivingAccountService $service): int
    {
        $snapshot = $service->primeRuntimeStoresFromDatabase();

        $this->info(sprintf(
            'Da prime %d tai khoan nhan tien vao cache va Redis chia se.',
            count($snapshot['accounts'] ?? []),
        ));

        return self::SUCCESS;
    }
}
