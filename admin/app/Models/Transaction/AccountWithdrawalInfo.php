<?php

namespace App\Models\Transaction;

use App\Enum\Wallet\UnitTransaction;
use App\Models\User;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\Relations\HasMany;
use Illuminate\Database\Eloquent\SoftDeletes;

class AccountWithdrawalInfo extends Model
{
    use SoftDeletes;

    protected $fillable = [
        'user_id',
        'unit',
        'provider_code',
        'account_name',
        'account_number',
        'is_default',
    ];

    protected function casts(): array
    {
        return [
            'unit' => UnitTransaction::class,
            'is_default' => 'boolean',
        ];
    }

    public function user(): BelongsTo
    {
        return $this->belongsTo(User::class);
    }

    public function withdrawalRequests(): HasMany
    {
        return $this->hasMany(WithdrawalRequest::class);
    }
}
