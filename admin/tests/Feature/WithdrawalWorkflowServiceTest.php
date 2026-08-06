<?php

namespace Tests\Feature;

use App\Enum\Transaction\WithdrawalStatus;
use App\Models\Transaction\WithdrawalRequest;
use App\Models\User;
use App\Services\Admin\WithdrawalWorkflowService;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Schema;
use Illuminate\Validation\ValidationException;
use Tests\TestCase;

class WithdrawalWorkflowServiceTest extends TestCase
{
    protected function setUp(): void
    {
        parent::setUp();

        Schema::dropIfExists('wallet_ledger_entries');
        Schema::dropIfExists('withdrawal_requests');
        Schema::dropIfExists('wallets');

        Schema::create('wallets', function (Blueprint $table): void {
            $table->id();
            $table->unsignedBigInteger('user_id');
            $table->unsignedTinyInteger('unit')->default(1);
            $table->string('balance');
            $table->string('locked_balance');
            $table->unsignedTinyInteger('status')->default(1);
            $table->timestamps();
        });
        Schema::create('withdrawal_requests', function (Blueprint $table): void {
            $table->id();
            $table->unsignedBigInteger('user_id');
            $table->unsignedBigInteger('wallet_id');
            $table->unsignedBigInteger('account_withdrawal_info_id')->nullable();
            $table->unsignedTinyInteger('unit')->default(1);
            $table->string('amount');
            $table->string('fee')->default('0.00000000');
            $table->string('net_amount');
            $table->unsignedTinyInteger('status');
            $table->text('reason_rejected')->nullable();
            $table->unsignedBigInteger('reviewed_by')->nullable();
            $table->timestamp('reviewed_at')->nullable();
            $table->unsignedBigInteger('paid_by')->nullable();
            $table->timestamp('paid_at')->nullable();
            $table->string('transfer_reference')->nullable();
            $table->string('transfer_proof_path')->nullable();
            $table->text('admin_note')->nullable();
            $table->timestamps();
            $table->softDeletes();
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

    public function test_reject_uses_exact_decimal_arithmetic_for_wallet_and_ledger(): void
    {
        $request = $this->createRequest(
            balance: '12345678901234567890.12345678',
            lockedBalance: '10.00000000',
            amount: '1.12345678',
            status: WithdrawalStatus::PENDING,
        );

        app(WithdrawalWorkflowService::class)->reject($request, $this->actor(), 'invalid details');

        self::assertSame(
            '12345678901234567891.24691356',
            DB::table('wallets')->where('id', $request->wallet_id)->value('balance'),
        );
        self::assertSame(
            '8.87654322',
            DB::table('wallets')->where('id', $request->wallet_id)->value('locked_balance'),
        );
        self::assertSame('1.12345678', DB::table('wallet_ledger_entries')->value('amount'));
        self::assertSame(
            '12345678901234567891.24691356',
            DB::table('wallet_ledger_entries')->value('balance_after'),
        );
        self::assertSame(
            WithdrawalStatus::REJECTED->value,
            (int) DB::table('withdrawal_requests')->where('id', $request->id)->value('status'),
        );
    }

    public function test_reject_rolls_back_when_locked_balance_is_insufficient(): void
    {
        $request = $this->createRequest(
            balance: '100.00000000',
            lockedBalance: '0.99999999',
            amount: '1.00000000',
            status: WithdrawalStatus::PENDING,
        );

        $this->assertInsufficientLockedBalance(function () use ($request): void {
            app(WithdrawalWorkflowService::class)->reject($request, $this->actor());
        });

        $this->assertRequestAndWalletUnchanged($request, '100.00000000', '0.99999999', WithdrawalStatus::PENDING);
    }

    public function test_mark_paid_rolls_back_when_locked_balance_is_insufficient(): void
    {
        $request = $this->createRequest(
            balance: '100.00000000',
            lockedBalance: '0.99999999',
            amount: '1.00000000',
            status: WithdrawalStatus::APPROVED,
        );

        $this->assertInsufficientLockedBalance(function () use ($request): void {
            app(WithdrawalWorkflowService::class)->markPaid($request, $this->actor(), 'transfer-1');
        });

        $this->assertRequestAndWalletUnchanged($request, '100.00000000', '0.99999999', WithdrawalStatus::APPROVED);
    }

    private function createRequest(
        string $balance,
        string $lockedBalance,
        string $amount,
        WithdrawalStatus $status,
    ): WithdrawalRequest {
        $now = now();
        $walletID = DB::table('wallets')->insertGetId([
            'user_id' => 123456,
            'unit' => 1,
            'balance' => $balance,
            'locked_balance' => $lockedBalance,
            'status' => 1,
            'created_at' => $now,
            'updated_at' => $now,
        ]);
        $requestID = DB::table('withdrawal_requests')->insertGetId([
            'user_id' => 123456,
            'wallet_id' => $walletID,
            'unit' => 1,
            'amount' => $amount,
            'fee' => '0.00000000',
            'net_amount' => $amount,
            'status' => $status->value,
            'created_at' => $now,
            'updated_at' => $now,
        ]);

        return WithdrawalRequest::query()->findOrFail($requestID);
    }

    private function actor(): User
    {
        $actor = new User;
        $actor->setAttribute('id', 654321);
        $actor->exists = true;

        return $actor;
    }

    private function assertInsufficientLockedBalance(callable $operation): void
    {
        try {
            $operation();
            self::fail('Expected insufficient locked balance validation error.');
        } catch (ValidationException $exception) {
            self::assertArrayHasKey('locked_balance', $exception->errors());
        }
    }

    private function assertRequestAndWalletUnchanged(
        WithdrawalRequest $request,
        string $balance,
        string $lockedBalance,
        WithdrawalStatus $status,
    ): void {
        self::assertSame($balance, DB::table('wallets')->where('id', $request->wallet_id)->value('balance'));
        self::assertSame($lockedBalance, DB::table('wallets')->where('id', $request->wallet_id)->value('locked_balance'));
        self::assertSame($status->value, (int) DB::table('withdrawal_requests')->where('id', $request->id)->value('status'));
        self::assertSame(0, DB::table('wallet_ledger_entries')->count());
    }
}
