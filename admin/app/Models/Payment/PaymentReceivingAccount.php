<?php

namespace App\Models\Payment;

use App\Enum\Payment\PaymentReceivingAccountStatus;
use App\Enum\Payment\PaymentReceivingAccountType;
use App\Enum\Wallet\UnitTransaction;
use App\Models\Transaction\Transaction;
use Illuminate\Database\Eloquent\Builder;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\Relations\HasMany;
use Illuminate\Database\Eloquent\SoftDeletes;

class PaymentReceivingAccount extends Model
{
    use SoftDeletes;

    protected $fillable = [
        'type',
        'unit',
        'provider_code',
        'account_name',
        'account_number',
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

    public function scopeActive(Builder $query): Builder
    {
        return $query->where('status', PaymentReceivingAccountStatus::ACTIVE->value);
    }

    public function scopeSepayVietQr(Builder $query): Builder
    {
        return $query
            ->where('type', PaymentReceivingAccountType::BANK->value)
            ->where('unit', UnitTransaction::VND->value);
    }

    public function scopeForDeposit(Builder $query): Builder
    {
        return $query
            ->active()
            ->sepayVietQr()
            ->orderByDesc('is_default')
            ->orderBy('sort_order')
            ->orderBy('id');
    }

    public function transactions(): HasMany
    {
        return $this->hasMany(Transaction::class, 'receiving_account_id');
    }

    public function bank(): BelongsTo
    {
        return $this->belongsTo(VietQrBank::class, 'provider_code', 'code');
    }
}
