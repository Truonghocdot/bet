<?php

namespace Tests\Feature;

use App\Enum\Transaction\TransactionStatus;
use App\Enum\Transaction\TypeTransaction;
use App\Enum\Transaction\WithdrawalStatus;
use App\Enum\Wallet\UnitTransaction;
use App\Enum\Wallet\WalletStatus;
use App\Models\Finance\FakeDepositTransaction;
use App\Models\Finance\FakeWithdrawTransaction;
use App\Models\Transaction\AccountWithdrawalInfo;
use App\Models\Transaction\Transaction;
use App\Models\Transaction\WithdrawalRequest;
use App\Models\User;
use App\Models\Wallet\Wallet;
use Illuminate\Support\Facades\Artisan;
use Tests\TestCase;

class FakeFinanceFeedHooksTest extends TestCase
{
    protected function setUp(): void
    {
        parent::setUp();

        $this->migrateFinanceTables();
    }

    public function test_completed_real_deposit_appends_one_fake_deposit_without_duplicates(): void
    {
        $user = User::factory()->create();
        $wallet = $this->createWallet($user);

        $transaction = Transaction::query()->create([
            'user_id' => $user->id,
            'wallet_id' => $wallet->id,
            'unit' => UnitTransaction::VND,
            'type' => TypeTransaction::DEPOSIT,
            'amount' => '150000.00000000',
            'fee' => '0.00000000',
            'net_amount' => '150000.00000000',
            'status' => TransactionStatus::PENDING,
        ]);

        $this->assertDatabaseCount('fake_deposit_transactions', 0);

        $transaction->update([
            'status' => TransactionStatus::COMPLETED,
            'approved_at' => now(),
        ]);

        $this->assertDatabaseCount('fake_deposit_transactions', 1);

        $fakeTransaction = FakeDepositTransaction::query()->sole();

        $this->assertSame('real_transaction_completed', $fakeTransaction->meta['trigger'] ?? null);
        $this->assertSame(Transaction::class, $fakeTransaction->meta['reference_type'] ?? null);
        $this->assertSame($transaction->id, $fakeTransaction->meta['reference_id'] ?? null);

        $transaction->update([
            'approved_at' => now()->addMinute(),
        ]);

        $this->assertDatabaseCount('fake_deposit_transactions', 1);
    }

    public function test_paid_withdrawal_request_appends_one_fake_withdraw_without_duplicates(): void
    {
        $user = User::factory()->create();
        $wallet = $this->createWallet($user);
        $withdrawalInfo = AccountWithdrawalInfo::query()->create([
            'user_id' => $user->id,
            'unit' => UnitTransaction::VND,
            'provider_code' => 'vcb',
            'account_name' => 'Test User',
            'account_number' => '0123456789',
            'is_default' => true,
        ]);

        $request = WithdrawalRequest::query()->create([
            'user_id' => $user->id,
            'wallet_id' => $wallet->id,
            'account_withdrawal_info_id' => $withdrawalInfo->id,
            'unit' => UnitTransaction::VND,
            'amount' => '120000.00000000',
            'fee' => '0.00000000',
            'net_amount' => '120000.00000000',
            'status' => WithdrawalStatus::APPROVED,
        ]);

        $this->assertDatabaseCount('fake_withdraw_transactions', 0);

        $request->update([
            'status' => WithdrawalStatus::PAID,
            'transfer_reference' => 'WD-TEST-001',
            'paid_at' => now(),
        ]);

        $this->assertDatabaseCount('fake_withdraw_transactions', 1);

        $fakeTransaction = FakeWithdrawTransaction::query()->sole();

        $this->assertSame('real_withdrawal_paid', $fakeTransaction->meta['trigger'] ?? null);
        $this->assertSame(WithdrawalRequest::class, $fakeTransaction->meta['reference_type'] ?? null);
        $this->assertSame($request->id, $fakeTransaction->meta['reference_id'] ?? null);

        $request->update([
            'paid_at' => now()->addMinute(),
        ]);

        $this->assertDatabaseCount('fake_withdraw_transactions', 1);
    }

    private function createWallet(User $user): Wallet
    {
        return Wallet::query()->create([
            'user_id' => $user->id,
            'unit' => UnitTransaction::VND,
            'balance' => '0.00000000',
            'locked_balance' => '0.00000000',
            'status' => WalletStatus::ACTIVE,
        ]);
    }

    private function migrateFinanceTables(): void
    {
        Artisan::call('db:wipe', ['--force' => true]);

        foreach ($this->migrationPaths() as $path) {
            Artisan::call('migrate', [
                '--force' => true,
                '--path' => $path,
                '--realpath' => true,
            ]);
        }
    }

    /**
     * @return list<string>
     */
    private function migrationPaths(): array
    {
        return [
            base_path('database/migrations/0001_01_01_000000_create_users_table.php'),
            base_path('database/migrations/2026_04_09_000002_create_wallets_table.php'),
            base_path('database/migrations/2026_04_09_000004_create_transactions_table.php'),
            base_path('database/migrations/2026_04_09_000005_create_account_withdrawal_infos_table.php'),
            base_path('database/migrations/2026_04_09_000006_create_withdrawal_requests_table.php'),
            base_path('database/migrations/2026_04_09_000017_add_manual_payout_fields_to_withdrawal_requests_table.php'),
            base_path('database/migrations/2026_06_18_000002_create_fake_finance_transaction_tables.php'),
        ];
    }
}
