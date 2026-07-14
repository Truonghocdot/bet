<?php

namespace App\Console\Commands;

use App\Support\FakeFinance\FakeFinanceFeedService;
use Illuminate\Console\Command;

class GenerateFakeFinanceTransactionsCommand extends Command
{
    protected $signature = 'finance:generate-fake-transactions';

    protected $description = 'Generate fake deposit and withdraw transaction feeds from the provided sample pool';

    public function handle(FakeFinanceFeedService $feedService): int
    {
        $depositInserted = $feedService->appendDepositBatch(random_int(1, 3), [
            'trigger' => 'scheduled_command',
        ]);
        $withdrawInserted = $feedService->appendWithdrawBatch(random_int(1, 3), [
            'trigger' => 'scheduled_command',
        ]);

        $this->info(sprintf(
            'Da them %d giao dich nap fake va %d giao dich rut fake.',
            $depositInserted,
            $withdrawInserted,
        ));

        return self::SUCCESS;
    }
}
