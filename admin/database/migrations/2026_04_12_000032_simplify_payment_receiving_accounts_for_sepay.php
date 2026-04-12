<?php

use App\Enum\Payment\PaymentReceivingAccountType;
use App\Enum\Wallet\UnitTransaction;
use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        DB::table('payment_receiving_accounts')->update([
            'type' => PaymentReceivingAccountType::BANK->value,
            'unit' => UnitTransaction::VND->value,
        ]);

        DB::statement('alter table payment_receiving_accounts alter column type set default 1');
        DB::statement('alter table payment_receiving_accounts alter column unit set default 1');

        DB::statement('drop index if exists payment_receiving_accounts_type_unit_status_index');

        Schema::table('payment_receiving_accounts', function (Blueprint $table): void {
            $columnsToDrop = ['wallet_address', 'network', 'qr_code_path', 'name', 'instructions', 'code'];
            foreach ($columnsToDrop as $column) {
                if (Schema::hasColumn('payment_receiving_accounts', $column)) {
                    $table->dropColumn($column);
                }
            }
        });

        DB::statement('create index if not exists idx_payment_receiving_accounts_provider_status on payment_receiving_accounts (provider_code, status)');
    }

    public function down(): void
    {
        Schema::table('payment_receiving_accounts', function (Blueprint $table): void {
            if (! Schema::hasColumn('payment_receiving_accounts', 'wallet_address')) {
                $table->string('wallet_address', 255)->nullable()->after('account_number');
            }

            if (! Schema::hasColumn('payment_receiving_accounts', 'network')) {
                $table->string('network', 50)->nullable()->after('wallet_address');
            }

            if (! Schema::hasColumn('payment_receiving_accounts', 'qr_code_path')) {
                $table->string('qr_code_path', 255)->nullable()->after('network');
            }

            if (! Schema::hasColumn('payment_receiving_accounts', 'name')) {
                $table->string('name', 100)->nullable()->after('qr_code_path');
            }

            if (! Schema::hasColumn('payment_receiving_accounts', 'instructions')) {
                $table->text('instructions')->nullable()->after('name');
            }

            if (! Schema::hasColumn('payment_receiving_accounts', 'code')) {
                $table->string('code', 50)->nullable()->after('id');
            }
        });

        DB::statement('drop index if exists idx_payment_receiving_accounts_provider_status');
        DB::statement('create index if not exists payment_receiving_accounts_type_unit_status_index on payment_receiving_accounts (type, unit, status)');
    }
};

