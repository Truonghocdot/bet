<?php

namespace App\Models\Transaction;

use App\Enum\Transaction\WithdrawalStatus;
use App\Enum\Wallet\UnitTransaction;
use App\Models\User;
use App\Models\Wallet\Wallet;
use Illuminate\Database\Eloquent\Casts\Attribute;
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
        'paid_by',
        'paid_at',
        'transfer_reference',
        'transfer_proof_path',
        'admin_note',
    ];

    protected function casts(): array
    {
        return [
            'unit' => UnitTransaction::class,
            'amount' => 'decimal:8',
            'fee' => 'decimal:8',
            'net_amount' => 'decimal:8',
            'reviewed_at' => 'datetime',
            'paid_at' => 'datetime',
        ];
    }

    protected function status(): Attribute
    {
        return Attribute::make(
            get: static fn (mixed $value): WithdrawalStatus => WithdrawalStatus::tryFrom((int) $value) ?? WithdrawalStatus::PENDING,
            set: static fn (mixed $value): int => $value instanceof WithdrawalStatus
                ? $value->value
                : (WithdrawalStatus::tryFrom((int) $value)?->value ?? WithdrawalStatus::PENDING->value),
        );
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

    public function paidBy(): BelongsTo
    {
        return $this->belongsTo(User::class, 'paid_by');
    }
}
