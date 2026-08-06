<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    private const SCALE = 8;

    private const CUMULATIVE_COLUMNS = [
        'wallets' => [
            'balance',
            'locked_balance',
        ],
        'wallet_ledger_entries' => [
            'balance_before',
            'balance_after',
        ],
    ];

    private const MONETARY_COLUMNS = [
        'wallet_ledger_entries' => [
            'amount',
        ],
        'transactions' => [
            'amount',
            'original_amount',
            'fee',
            'net_amount',
        ],
        'withdrawal_requests' => [
            'amount',
            'fee',
            'net_amount',
        ],
        'bet_tickets' => [
            'stake',
            'total_stake',
            'original_amount',
            'tax_amount',
            'net_amount',
            'potential_payout',
            'actual_payout',
        ],
        'bet_items' => [
            'stake',
            'payout_amount',
        ],
        'bet_settlements' => [
            'payout_amount',
            'profit_loss',
        ],
        'affiliate_referrals' => [
            'first_deposit_amount',
        ],
        'affiliate_reward_settings' => [
            'reward_amount',
        ],
        'affiliate_reward_logs' => [
            'reward_amount',
        ],
        'exchange_rate_settings' => [
            'withdraw_required_bet_volume',
            'withdraw_min_amount',
            'withdraw_max_amount',
        ],
    ];

    public function up(): void
    {
        $this->changeColumnTypes(self::CUMULATIVE_COLUMNS, 'NUMERIC');
        $this->changeColumnTypes(
            self::MONETARY_COLUMNS,
            sprintf('NUMERIC(%d,%d)', 30, self::SCALE),
        );
    }

    public function down(): void
    {
        // Rollback requires every value to fit NUMERIC(20,8); PostgreSQL will abort otherwise.
        $legacyType = sprintf('NUMERIC(%d,%d)', 20, self::SCALE);
        $this->changeColumnTypes(self::MONETARY_COLUMNS, $legacyType);
        $this->changeColumnTypes(self::CUMULATIVE_COLUMNS, $legacyType);
    }

    private function changeColumnTypes(array $columnsByTable, string $numericType): void
    {
        // SQLite does not enforce DECIMAL precision, while the runtime engine uses PostgreSQL.
        if (DB::getDriverName() !== 'pgsql') {
            return;
        }

        foreach ($columnsByTable as $table => $columns) {
            if (! Schema::hasTable($table)) {
                continue;
            }

            foreach ($columns as $column) {
                if (! Schema::hasColumn($table, $column)) {
                    continue;
                }

                $quotedTable = $this->quoteIdentifier($table);
                $quotedColumn = $this->quoteIdentifier($column);

                DB::statement(sprintf(
                    'ALTER TABLE %s ALTER COLUMN %s TYPE %s USING %s::%s',
                    $quotedTable,
                    $quotedColumn,
                    $numericType,
                    $quotedColumn,
                    $numericType,
                ));
            }
        }
    }

    private function quoteIdentifier(string $identifier): string
    {
        if (preg_match('/\A[a-z_][a-z0-9_]*\z/', $identifier) !== 1) {
            throw new InvalidArgumentException('Invalid PostgreSQL identifier.');
        }

        return '"'.str_replace('"', '""', $identifier).'"';
    }
};
