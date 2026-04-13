<?php

use App\Enum\Transaction\WithdrawalStatus;
use Illuminate\Database\Migrations\Migration;
use Illuminate\Support\Facades\DB;

return new class extends Migration
{
    public function up(): void
    {
        DB::table('withdrawal_requests')
            ->whereNotIn('status', array_map(static fn (WithdrawalStatus $status): int => $status->value, WithdrawalStatus::cases()))
            ->update(['status' => WithdrawalStatus::PENDING->value]);

        DB::statement('ALTER TABLE withdrawal_requests ALTER COLUMN status SET DEFAULT '.WithdrawalStatus::PENDING->value);
        DB::statement('ALTER TABLE withdrawal_requests DROP CONSTRAINT IF EXISTS withdrawal_requests_status_check');
        DB::statement(sprintf(
            'ALTER TABLE withdrawal_requests ADD CONSTRAINT withdrawal_requests_status_check CHECK (status IN (%s))',
            implode(',', array_map(static fn (WithdrawalStatus $status): int => $status->value, WithdrawalStatus::cases()))
        ));
    }

    public function down(): void
    {
        DB::statement('ALTER TABLE withdrawal_requests DROP CONSTRAINT IF EXISTS withdrawal_requests_status_check');
        DB::statement('ALTER TABLE withdrawal_requests ALTER COLUMN status DROP DEFAULT');
    }
};

