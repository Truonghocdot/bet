<?php

namespace App\Models\Wallet;

use App\Enum\Wallet\LedgerDirection;
use App\Models\User;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

class WalletLedgerEntry extends Model
{
    public $timestamps = false;

    protected $fillable = [
        'wallet_id',
        'user_id',
        'direction',
        'amount',
        'balance_before',
        'balance_after',
        'reference_type',
        'reference_id',
        'note',
        'created_at',
    ];

    protected function casts(): array
    {
        return [
            'direction' => LedgerDirection::class,
            'amount' => 'decimal:8',
            'balance_before' => 'decimal:8',
            'balance_after' => 'decimal:8',
            'created_at' => 'datetime',
        ];
    }

    public function wallet(): BelongsTo
    {
        return $this->belongsTo(Wallet::class);
    }

    public function user(): BelongsTo
    {
        return $this->belongsTo(User::class);
    }
}
