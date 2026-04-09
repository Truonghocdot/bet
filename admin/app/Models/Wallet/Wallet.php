<?php

namespace App\Models\Wallet;

use App\Enum\Wallet\UnitTransaction;
use App\Enum\Wallet\WalletStatus;
use App\Models\Bet\BetTicket;
use App\Models\Transaction\Transaction;
use App\Models\Transaction\WithdrawalRequest;
use App\Models\User;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\Relations\HasMany;

class Wallet extends Model
{
    protected $fillable = [
        'user_id',
        'unit',
        'balance',
        'locked_balance',
        'status',
    ];

    protected function casts(): array
    {
        return [
            'unit' => UnitTransaction::class,
            'balance' => 'decimal:8',
            'locked_balance' => 'decimal:8',
            'status' => WalletStatus::class,
        ];
    }

    public function user(): BelongsTo
    {
        return $this->belongsTo(User::class);
    }

    public function ledgerEntries(): HasMany
    {
        return $this->hasMany(WalletLedgerEntry::class);
    }

    public function transactions(): HasMany
    {
        return $this->hasMany(Transaction::class);
    }

    public function withdrawalRequests(): HasMany
    {
        return $this->hasMany(WithdrawalRequest::class);
    }

    public function betTickets(): HasMany
    {
        return $this->hasMany(BetTicket::class);
    }
}
