<?php

namespace App\Models\Transaction;

use App\Enum\Transaction\TransactionStatus;
use App\Enum\Transaction\TypeTransaction;
use App\Enum\Wallet\UnitTransaction;
use App\Models\User;
use App\Models\Payment\PaymentReceivingAccount;
use App\Models\Wallet\Wallet;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\SoftDeletes;

class Transaction extends Model
{
    use SoftDeletes;

    protected $fillable = [
        'user_id',
        'wallet_id',
        'client_ref',
        'unit',
        'type',
        'amount',
        'fee',
        'net_amount',
        'status',
        'provider',
        'provider_txn_id',
        'receiving_account_id',
        'meta',
        'reason_failed',
        'approved_by',
        'approved_at',
    ];

    protected function casts(): array
    {
        return [
            'unit' => UnitTransaction::class,
            'type' => TypeTransaction::class,
            'amount' => 'decimal:8',
            'fee' => 'decimal:8',
            'net_amount' => 'decimal:8',
            'status' => TransactionStatus::class,
            'approved_at' => 'datetime',
            'meta' => 'array',
        ];
    }

    public function user(): BelongsTo
    {
        return $this->belongsTo(User::class);
    }

    public function wallet(): BelongsTo
    {
        return $this->belongsTo(Wallet::class);
    }

    public function receivingAccount(): BelongsTo
    {
        return $this->belongsTo(PaymentReceivingAccount::class, 'receiving_account_id');
    }

    public function approvedBy(): BelongsTo
    {
        return $this->belongsTo(User::class, 'approved_by');
    }
}
