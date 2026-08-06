<?php

namespace App\Services\Admin;

use App\Enum\Transaction\TransactionStatus;
use App\Enum\Wallet\LedgerDirection;
use App\Models\Transaction\Transaction;
use App\Models\Wallet\WalletLedgerEntry;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Auth;

class DepositWorkflowService
{
    /**
     * Phê duyệt lệnh nạp tiền
     */
    public function approve(Transaction $transaction): bool
    {
        return DB::transaction(function () use ($transaction) {
            $transaction = Transaction::query()
                ->lockForUpdate()
                ->findOrFail($transaction->getKey());

            if ($transaction->status !== TransactionStatus::PENDING) {
                return false;
            }

            // Khóa ví để cập nhật số dư an toàn
            $wallet = DB::table('wallets')
                ->where('id', $transaction->wallet_id)
                ->lockForUpdate()
                ->first();
            if (! $wallet) {
                return false;
            }

            $balanceBefore = $wallet->balance;
            $balanceAfter = bcadd($balanceBefore, $transaction->amount, 8);

            // 1. Cập nhật số dư ví
            DB::table('wallets')
                ->where('id', $wallet->id)
                ->update([
                    'balance' => $balanceAfter,
                    'updated_at' => now(),
                ]);

            // 2. Ghi log biến động số dư
            WalletLedgerEntry::create([
                'wallet_id' => $wallet->id,
                'user_id' => $transaction->user_id,
                'direction' => LedgerDirection::CREDIT,
                'amount' => $transaction->amount,
                'balance_before' => $balanceBefore,
                'balance_after' => $balanceAfter,
                'reference_type' => Transaction::class,
                'reference_id' => $transaction->id,
                'note' => 'Hệ thống duyệt nạp tiền thủ công',
                'created_at' => now(),
            ]);

            // 3. Cập nhật trạng thái giao dịch
            $transaction->update([
                'status' => TransactionStatus::COMPLETED,
                'approved_by' => Auth::id(),
                'approved_at' => now(),
            ]);

            return true;
        });
    }

    /**
     * Từ chối lệnh nạp tiền
     */
    public function reject(Transaction $transaction, ?string $reason = null): bool
    {
        return DB::transaction(function () use ($transaction, $reason): bool {
            $transaction = Transaction::query()
                ->lockForUpdate()
                ->findOrFail($transaction->getKey());

            if ($transaction->status !== TransactionStatus::PENDING) {
                return false;
            }

            return $transaction->update([
                'status' => TransactionStatus::FAILED,
                'reason_failed' => $reason ?? 'Quản trị viên từ chối lệnh nạp',
                'approved_by' => Auth::id(),
                'approved_at' => now(),
            ]);
        });
    }
}
