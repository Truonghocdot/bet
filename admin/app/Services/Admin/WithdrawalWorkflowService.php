<?php

namespace App\Services\Admin;

use App\Enum\Transaction\TransactionStatus;
use App\Enum\Transaction\TypeTransaction;
use App\Enum\Transaction\WithdrawalStatus;
use App\Enum\Wallet\LedgerDirection;
use App\Enum\Wallet\WalletStatus;
use App\Models\Transaction\Transaction;
use App\Models\Transaction\WithdrawalRequest;
use App\Models\Wallet\WalletLedgerEntry;
use App\Models\User;
use Illuminate\Support\Facades\DB;
use Illuminate\Validation\ValidationException;

class WithdrawalWorkflowService
{
    public function approve(WithdrawalRequest $request, User $actor): WithdrawalRequest
    {
        return DB::transaction(function () use ($request, $actor): WithdrawalRequest {
            $request->refresh();

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
            $request->refresh();

            if (in_array($request->status, [WithdrawalStatus::PAID, WithdrawalStatus::CANCELED], true)) {
                throw ValidationException::withMessages([
                    'status' => 'Yêu cầu rút đã được xử lý xong, không thể từ chối.',
                ]);
            }

            $wallet = $request->wallet()->lockForUpdate()->first();
            if ($wallet) {
                $balanceBefore = $wallet->balance;
                $lockedBefore = $wallet->locked_balance;

                // Hoàn tiền: Cộng lại vào balance khả dụng, trừ ở tiền khóa
                $wallet->forceFill([
                    'balance' => $balanceBefore + $request->amount,
                    'locked_balance' => max(0, $lockedBefore - $request->amount),
                ])->save();

                WalletLedgerEntry::create([
                    'wallet_id' => $wallet->id,
                    'user_id' => $request->user_id,
                    'direction' => LedgerDirection::CREDIT,
                    'amount' => $request->amount,
                    'balance_before' => $balanceBefore,
                    'balance_after' => $balanceBefore + $request->amount,
                    'reference_type' => WithdrawalRequest::class,
                    'reference_id' => $request->id,
                    'note' => 'Hoàn tiền do từ chối yêu cầu rút',
                    'created_at' => now(),
                ]);
            }

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
            $request->refresh();

            if ($request->status !== WithdrawalStatus::APPROVED) {
                throw ValidationException::withMessages([
                    'status' => 'Yêu cầu rút phải được duyệt trước khi xác nhận đã chi trả.',
                ]);
            }

            $wallet = $request->wallet()->lockForUpdate()->first();
            if ($wallet) {
                $balanceBefore = $wallet->balance;
                $lockedBefore = $wallet->locked_balance;

                // Khi chi trả: Tiền khóa "bay màu", balance khả dụng giữ nguyên (vì đã trừ lúc tạo lệnh)
                $wallet->forceFill([
                    'locked_balance' => max(0, $lockedBefore - $request->amount),
                ])->save();

                WalletLedgerEntry::create([
                    'wallet_id' => $wallet->id,
                    'user_id' => $request->user_id,
                    'direction' => LedgerDirection::NEUTRAL, // Trung tính vì balance khả dụng không đổi thêm
                    'amount' => $request->amount,
                    'balance_before' => $balanceBefore,
                    'balance_after' => $balanceBefore,
                    'reference_type' => WithdrawalRequest::class,
                    'reference_id' => $request->id,
                    'note' => 'Giải phóng tiền khóa: Đã chi trả xong lệnh rút',
                    'created_at' => now(),
                ]);
            }

            Transaction::create([
                'user_id' => $request->user_id,
                'wallet_id' => $request->wallet_id,
                'unit' => $request->unit,
                'type' => TypeTransaction::WITHDRAW,
                'amount' => $request->amount,
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
}
