<?php

namespace App\Console\Commands;

use App\Enum\Transaction\TransactionStatus;
use App\Enum\Transaction\TypeTransaction;
use App\Models\Transaction\Transaction;
use Illuminate\Console\Command;

class ExpirePendingDepositsCommand extends Command
{
    protected $signature = 'deposits:expire-pending';

    protected $description = 'Auto cancel pending deposit transactions older than 3 minutes';

    public function handle(): int
    {
        $cutoffAt = now()->subMinutes(3);
        $expiredCount = 0;

        Transaction::query()
            ->where('type', TypeTransaction::DEPOSIT->value)
            ->where('status', TransactionStatus::PENDING->value)
            ->where('created_at', '<=', $cutoffAt)
            ->chunkById(200, function ($transactions) use (&$expiredCount): void {
                foreach ($transactions as $transaction) {
                    $transaction->forceFill([
                        'status' => TransactionStatus::CANCELED,
                        'reason_failed' => 'Hệ thống tự hủy lệnh nạp do đã hết hạn sau 3 phút.',
                    ])->save();

                    $expiredCount++;
                }
            });

        $this->info(sprintf(
            'Da tu dong huy %d lenh nap pending qua han truoc %s.',
            $expiredCount,
            $cutoffAt->format('Y-m-d H:i:s'),
        ));

        return self::SUCCESS;
    }
}
