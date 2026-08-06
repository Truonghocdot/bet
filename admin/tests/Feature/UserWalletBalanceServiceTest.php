<?php

namespace Tests\Feature;

use App\Enum\Wallet\UnitTransaction;
use App\Models\User;
use App\Services\Admin\UserWalletBalanceService;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Schema;
use Illuminate\Validation\ValidationException;
use Tests\TestCase;

class UserWalletBalanceServiceTest extends TestCase
{
    protected function setUp(): void
    {
        parent::setUp();

        Schema::dropIfExists('wallet_ledger_entries');
        Schema::dropIfExists('wallets');

        Schema::create('wallets', function (Blueprint $table): void {
            $table->id();
            $table->unsignedBigInteger('user_id');
            $table->unsignedTinyInteger('unit');
            $table->string('balance');
            $table->string('locked_balance');
            $table->unsignedTinyInteger('status');
            $table->timestamps();
        });
        Schema::create('wallet_ledger_entries', function (Blueprint $table): void {
            $table->id();
            $table->unsignedBigInteger('wallet_id');
            $table->unsignedBigInteger('user_id');
            $table->unsignedTinyInteger('direction');
            $table->string('amount');
            $table->string('balance_before');
            $table->string('balance_after');
            $table->string('reference_type');
            $table->unsignedBigInteger('reference_id')->nullable();
            $table->string('note')->nullable();
            $table->timestamp('created_at');
        });
    }

    public function test_sync_available_balance_preserves_large_decimal_values_exactly(): void
    {
        $user = $this->user();
        $walletID = $this->createWallet($user, '10000000000000000000.00000000');

        app(UserWalletBalanceService::class)->syncAvailableBalances($user, [
            UnitTransaction::VND->value => '12345678901234567890.12345678',
        ]);

        self::assertSame(
            '12345678901234567890.12345678',
            DB::table('wallets')->where('id', $walletID)->value('balance'),
        );
        self::assertSame(
            '2345678901234567890.12345678',
            DB::table('wallet_ledger_entries')->value('amount'),
        );
        self::assertSame(
            '10000000000000000000.00000000',
            DB::table('wallet_ledger_entries')->value('balance_before'),
        );
        self::assertSame(
            '12345678901234567890.12345678',
            DB::table('wallet_ledger_entries')->value('balance_after'),
        );
    }

    public function test_adjust_available_balance_preserves_the_eighth_decimal_place(): void
    {
        $user = $this->user();
        $walletID = $this->createWallet($user, '99999999999999999999.99999998');

        $delta = app(UserWalletBalanceService::class)->adjustAvailableBalance(
            $user,
            UnitTransaction::VND->value,
            '0.00000001',
        );

        self::assertSame('0.00000001', $delta);
        self::assertSame(
            '99999999999999999999.99999999',
            DB::table('wallets')->where('id', $walletID)->value('balance'),
        );
        self::assertSame('0.00000001', DB::table('wallet_ledger_entries')->value('amount'));
    }

    public function test_sync_rejects_float_input_instead_of_rounding_it(): void
    {
        $user = $this->user();
        $walletID = $this->createWallet($user, '100.00000000');

        try {
            app(UserWalletBalanceService::class)->syncAvailableBalances($user, [
                UnitTransaction::VND->value => 100.00000001,
            ]);
            self::fail('Expected float input validation error.');
        } catch (ValidationException $exception) {
            self::assertArrayHasKey('balance', $exception->errors());
        }

        self::assertSame('100.00000000', DB::table('wallets')->where('id', $walletID)->value('balance'));
        self::assertSame(0, DB::table('wallet_ledger_entries')->count());
    }

    private function user(): User
    {
        $user = new User;
        $user->setAttribute('id', 123456);
        $user->exists = true;

        return $user;
    }

    private function createWallet(User $user, string $balance): int
    {
        $now = now();

        return DB::table('wallets')->insertGetId([
            'user_id' => $user->id,
            'unit' => UnitTransaction::VND->value,
            'balance' => $balance,
            'locked_balance' => '0.00000000',
            'status' => 1,
            'created_at' => $now,
            'updated_at' => $now,
        ]);
    }
}
