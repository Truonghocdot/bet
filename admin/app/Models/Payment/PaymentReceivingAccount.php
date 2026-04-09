<?php

namespace App\Models\Payment;

use App\Enum\Payment\PaymentReceivingAccountStatus;
use App\Enum\Payment\PaymentReceivingAccountType;
use App\Enum\Wallet\UnitTransaction;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\SoftDeletes;

class PaymentReceivingAccount extends Model
{
    use SoftDeletes;

    protected $fillable = [
        'code',
        'name',
        'type',
        'unit',
        'provider_code',
        'account_name',
        'account_number',
        'wallet_address',
        'network',
        'qr_code_path',
        'instructions',
        'status',
        'is_default',
        'sort_order',
    ];

    protected function casts(): array
    {
        return [
            'type' => PaymentReceivingAccountType::class,
            'unit' => UnitTransaction::class,
            'status' => PaymentReceivingAccountStatus::class,
            'is_default' => 'boolean',
            'sort_order' => 'integer',
        ];
    }
}
