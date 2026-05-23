<?php

namespace App\Services\Admin;

use App\Enum\Wallet\LedgerDirection;
use App\Enum\Wallet\UnitTransaction;
use App\Enum\Wallet\WalletStatus;
use App\Models\User;
use App\Models\Wallet\Wallet;
use App\Models\Wallet\WalletLedgerEntry;
use Illuminate\Support\Facades\DB;

class UserWalletBalanceService
{
    /**
     * @param  array<int, mixed>  $balances
     */
    public function syncAvailableBalances(User $user, array $balances, ?User $actor = null): void
    {
        if ($balances === []) {
            return;
        }

        DB::transaction(function () use ($user, $balances, $actor): void {
            foreach ($balances as $unit => $targetValue) {
                $targetBalance = $this->normalizeDecimal($targetValue);

                $wallet = Wallet::query()
                    ->where('user_id', $user->id)
                    ->where('unit', $unit)
                    ->lockForUpdate()
                    ->first();

                if (! $wallet) {
                    $wallet = Wallet::query()->create([
                        'user_id' => $user->id,
                        'unit' => $unit,
                        'balance' => '0.00000000',
                        'locked_balance' => '0.00000000',
                        'status' => WalletStatus::ACTIVE,
                    ]);

                    $wallet = Wallet::query()
                        ->whereKey($wallet->id)
                        ->lockForUpdate()
                        ->firstOrFail();
                }

                $balanceBefore = $this->normalizeDecimal($wallet->balance);

                if (bccomp($balanceBefore, $targetBalance, 8) === 0) {
                    continue;
                }

                $wallet->forceFill([
                    'balance' => $targetBalance,
                    'status' => WalletStatus::ACTIVE,
                ])->save();

                $delta = bcsub($targetBalance, $balanceBefore, 8);

                WalletLedgerEntry::query()->create([
                    'wallet_id' => $wallet->id,
                    'user_id' => $user->id,
                    'direction' => bccomp($delta, '0.00000000', 8) >= 0
                        ? LedgerDirection::CREDIT
                        : LedgerDirection::DEBIT,
                    'amount' => $this->absoluteDecimal($delta),
                    'balance_before' => $balanceBefore,
                    'balance_after' => $targetBalance,
                    'reference_type' => User::class,
                    'reference_id' => $user->id,
                    'note' => $this->buildLedgerNote($unit, $actor),
                    'created_at' => now(),
                ]);
            }
        });
    }

    public function adjustAvailableBalance(User $user, int $unit, mixed $deltaValue, ?User $actor = null): string
    {
        $delta = $this->normalizeSignedDecimal($deltaValue);

        if (bccomp($delta, '0.00000000', 8) === 0) {
            return '0.00000000';
        }

        DB::transaction(function () use ($user, $unit, $delta, $actor): void {
            $wallet = Wallet::query()
                ->where('user_id', $user->id)
                ->where('unit', $unit)
                ->lockForUpdate()
                ->first();

            if (! $wallet) {
                $wallet = Wallet::query()->create([
                    'user_id' => $user->id,
                    'unit' => $unit,
                    'balance' => '0.00000000',
                    'locked_balance' => '0.00000000',
                    'status' => WalletStatus::ACTIVE,
                ]);

                $wallet = Wallet::query()
                    ->whereKey($wallet->id)
                    ->lockForUpdate()
                    ->firstOrFail();
            }

            $balanceBefore = $this->normalizeDecimal($wallet->balance);
            $balanceAfter = $this->normalizeDecimal(bcadd($balanceBefore, $delta, 8));

            $wallet->forceFill([
                'balance' => $balanceAfter,
                'status' => WalletStatus::ACTIVE,
            ])->save();

            WalletLedgerEntry::query()->create([
                'wallet_id' => $wallet->id,
                'user_id' => $user->id,
                'direction' => bccomp($delta, '0.00000000', 8) >= 0
                    ? LedgerDirection::CREDIT
                    : LedgerDirection::DEBIT,
                'amount' => $this->absoluteDecimal($delta),
                'balance_before' => $balanceBefore,
                'balance_after' => $balanceAfter,
                'reference_type' => User::class,
                'reference_id' => $user->id,
                'note' => $this->buildAdjustmentLedgerNote($unit, $delta, $actor),
                'created_at' => now(),
            ]);
        });

        return $delta;
    }

    private function buildLedgerNote(int $unit, ?User $actor): string
    {
        $unitLabel = match ((int) $unit) {
            UnitTransaction::USDT->value => 'USDT',
            default => 'VND',
        };

        $actorLabel = $actor
            ? trim(($actor->name ?: 'Admin #'.$actor->id).' (#'.$actor->id.')')
            : 'Hệ thống';

        return 'Điều chỉnh số dư ví '.$unitLabel.' từ form người dùng bởi '.$actorLabel;
    }

    private function buildAdjustmentLedgerNote(int $unit, string $delta, ?User $actor): string
    {
        $unitLabel = match ((int) $unit) {
            UnitTransaction::USDT->value => 'USDT',
            default => 'VND',
        };

        $actorLabel = $actor
            ? trim(($actor->name ?: 'Admin #'.$actor->id).' (#'.$actor->id.')')
            : 'Hệ thống';

        return sprintf(
            'Điều chỉnh nhanh ví %s %s bởi %s',
            $unitLabel,
            $delta,
            $actorLabel,
        );
    }

    private function normalizeDecimal(mixed $value): string
    {
        $normalized = str_replace([',', ' '], ['', ''], trim((string) $value));

        if ($normalized === '' || ! is_numeric($normalized)) {
            return '0.00000000';
        }

        return number_format((float) $normalized, 8, '.', '');
    }

    private function normalizeSignedDecimal(mixed $value): string
    {
        $normalized = str_replace([',', ' '], ['', ''], trim((string) $value));

        if ($normalized === '' || ! preg_match('/^[+-]?\d+(?:\.\d+)?$/', $normalized)) {
            return '0.00000000';
        }

        return number_format((float) $normalized, 8, '.', '');
    }

    private function absoluteDecimal(string $value): string
    {
        return ltrim($value, '-') ?: '0.00000000';
    }
}
