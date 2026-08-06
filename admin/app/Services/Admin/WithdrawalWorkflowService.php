<?php

namespace App\Services\Admin;

use App\Enum\Transaction\TransactionStatus;
use App\Enum\Transaction\TypeTransaction;
use App\Enum\Transaction\WithdrawalStatus;
use App\Enum\Wallet\LedgerDirection;
use App\Models\Transaction\Transaction;
use App\Models\Transaction\WithdrawalRequest;
use App\Models\User;
use App\Models\Wallet\WalletLedgerEntry;
use App\Support\Decimal;
use Illuminate\Support\Facades\DB;
use Illuminate\Validation\ValidationException;

class WithdrawalWorkflowService
{
    public function approve(WithdrawalRequest $request, User $actor): WithdrawalRequest
    {
        return DB::transaction(function () use ($request, $actor): WithdrawalRequest {
            $request = WithdrawalRequest::query()
                ->lockForUpdate()
                ->findOrFail($request->getKey());

            if ($request->status !== WithdrawalStatus::PENDING) {
                throw ValidationException::withMessages([
                    'status' => 'Yêu cầu rút không ở trạng thái chờ duyệt.',
                ]);
            }

            $request->forceFill([
                'status' => WithdrawalStatus::APPROVED,
                'reviewed_by' => $actor->id,
                'reviewed_at' => now(),
            ])->save();

            return $request->refresh();
        });
    }

    public function reject(WithdrawalRequest $request, User $actor, ?string $reason = null): WithdrawalRequest
    {
        return DB::transaction(function () use ($request, $actor, $reason): WithdrawalRequest {
            $request = WithdrawalRequest::query()
                ->lockForUpdate()
                ->findOrFail($request->getKey());

            if (! in_array($request->status, [WithdrawalStatus::PENDING, WithdrawalStatus::APPROVED], true)) {
                throw ValidationException::withMessages([
                    'status' => 'Yêu cầu rút đã được xử lý xong, không thể từ chối.',
                ]);
            }

            $wallet = $request->wallet()->lockForUpdate()->first();
            if (! $wallet) {
                throw ValidationException::withMessages([
                    'wallet' => 'Không tìm thấy ví của yêu cầu rút.',
                ]);
            }

            $balanceBefore = $this->normalizeDecimal($wallet->balance, 'balance');
            $lockedBefore = $this->normalizeDecimal($wallet->locked_balance, 'locked_balance');
            $amount = $this->normalizeDecimal($request->amount, 'amount');
            if (bccomp($lockedBefore, $amount, Decimal::SCALE) < 0) {
                throw ValidationException::withMessages([
                    'locked_balance' => 'Số dư khóa không đủ để từ chối và hoàn tiền yêu cầu rút.',
                ]);
            }

            $balanceAfter = bcadd($balanceBefore, $amount, Decimal::SCALE);
            $lockedAfter = bcsub($lockedBefore, $amount, Decimal::SCALE);
            $wallet->forceFill([
                'balance' => $balanceAfter,
                'locked_balance' => $lockedAfter,
            ])->save();

            WalletLedgerEntry::create([
                'wallet_id' => $wallet->id,
                'user_id' => $request->user_id,
                'direction' => LedgerDirection::CREDIT,
                'amount' => $amount,
                'balance_before' => $balanceBefore,
                'balance_after' => $balanceAfter,
                'reference_type' => WithdrawalRequest::class,
                'reference_id' => $request->id,
                'note' => 'Hoàn tiền do từ chối yêu cầu rút',
                'created_at' => now(),
            ]);

            $request->forceFill([
                'status' => WithdrawalStatus::REJECTED,
                'reason_rejected' => $reason,
                'reviewed_by' => $actor->id,
                'reviewed_at' => now(),
            ])->save();

            return $request->refresh();
        });
    }

    public function markPaid(
        WithdrawalRequest $request,
        User $actor,
        string $transferReference,
        ?string $proofPath = null,
    ): WithdrawalRequest {
        return DB::transaction(function () use ($request, $actor, $transferReference, $proofPath): WithdrawalRequest {
            $request = WithdrawalRequest::query()
                ->lockForUpdate()
                ->findOrFail($request->getKey());

            if ($request->status !== WithdrawalStatus::APPROVED) {
                throw ValidationException::withMessages([
                    'status' => 'Yêu cầu rút phải được duyệt trước khi xác nhận đã chi trả.',
                ]);
            }

            $wallet = $request->wallet()->lockForUpdate()->first();
            if (! $wallet) {
                throw ValidationException::withMessages([
                    'wallet' => 'Không tìm thấy ví của yêu cầu rút.',
                ]);
            }

            $balanceBefore = $this->normalizeDecimal($wallet->balance, 'balance');
            $lockedBefore = $this->normalizeDecimal($wallet->locked_balance, 'locked_balance');
            $amount = $this->normalizeDecimal($request->amount, 'amount');
            if (bccomp($lockedBefore, $amount, Decimal::SCALE) < 0) {
                throw ValidationException::withMessages([
                    'locked_balance' => 'Số dư khóa không đủ để xác nhận đã chi trả yêu cầu rút.',
                ]);
            }

            $lockedAfter = bcsub($lockedBefore, $amount, Decimal::SCALE);
            $wallet->forceFill([
                'locked_balance' => $lockedAfter,
            ])->save();

            WalletLedgerEntry::create([
                'wallet_id' => $wallet->id,
                'user_id' => $request->user_id,
                'direction' => LedgerDirection::NEUTRAL,
                'amount' => $amount,
                'balance_before' => $balanceBefore,
                'balance_after' => $balanceBefore,
                'reference_type' => WithdrawalRequest::class,
                'reference_id' => $request->id,
                'note' => 'Giải phóng tiền khóa: Đã chi trả xong lệnh rút',
                'created_at' => now(),
            ]);

            Transaction::create([
                'user_id' => $request->user_id,
                'wallet_id' => $request->wallet_id,
                'unit' => $request->unit,
                'type' => TypeTransaction::WITHDRAW,
                'amount' => $request->amount,
                'original_amount' => $request->amount,
                'fee' => $request->fee,
                'net_amount' => $request->net_amount,
                'status' => TransactionStatus::COMPLETED,
                'provider' => 'manual',
                'provider_txn_id' => $transferReference,
                'approved_by' => $actor->id,
                'approved_at' => now(),
            ]);

            $request->forceFill([
                'status' => WithdrawalStatus::PAID,
                'paid_by' => $actor->id,
                'paid_at' => now(),
                'transfer_reference' => $transferReference,
                'transfer_proof_path' => $proofPath,
            ])->save();

            return $request->refresh();
        });
    }

    private function normalizeDecimal(mixed $value, string $field): string
    {
        $normalized = Decimal::normalize($value);
        if ($normalized === null) {
            throw ValidationException::withMessages([
                $field => 'Giá trị tiền không hợp lệ hoặc có quá 8 chữ số thập phân.',
            ]);
        }

        return $normalized;
    }
}
