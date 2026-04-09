<?php

namespace App\Models\Transaction;

use App\Enum\Transaction\WithdrawalStatus;
use App\Enum\Wallet\UnitTransaction;
use App\Models\User;
use App\Models\Wallet\Wallet;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\SoftDeletes;

class WithdrawalRequest extends Model
{
    use SoftDeletes;

    protected $fillable = [
        'user_id',
        'wallet_id',
        'account_withdrawal_info_id',
        'unit',
        'amount',
        'fee',
        'net_amount',
        'status',
        'reason_rejected',
        'reviewed_by',
        'reviewed_at',
    ];

    protected function casts(): array
    {
        return [
            'unit' => UnitTransaction::class,
            'amount' => 'decimal:8',
            'fee' => 'decimal:8',
            'net_amount' => 'decimal:8',
            'status' => WithdrawalStatus::class,
            'reviewed_at' => 'datetime',
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

    public function accountWithdrawalInfo(): BelongsTo
    {
        return $this->belongsTo(AccountWithdrawalInfo::class);
    }

    public function reviewedBy(): BelongsTo
    {
        return $this->belongsTo(User::class, 'reviewed_by');
    }
}
