<?php

namespace App\Models\Wheel;

use App\Models\User;
use App\Models\Wallet\WalletLedgerEntry;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

class WheelReward extends Model
{
    protected $fillable = ['session_id', 'invitation_round_id', 'user_id', 'round_no', 'unit', 'amount', 'status', 'idempotency_key', 'wallet_ledger_entry_id', 'attempts', 'last_error', 'paid_at'];

    protected function casts(): array
    {
        return ['round_no' => 'integer', 'unit' => 'integer', 'amount' => 'decimal:8', 'attempts' => 'integer', 'paid_at' => 'datetime'];
    }

    public function session(): BelongsTo
    {
        return $this->belongsTo(WheelSession::class, 'session_id');
    }

    public function user(): BelongsTo
    {
        return $this->belongsTo(User::class);
    }

    public function ledgerEntry(): BelongsTo
    {
        return $this->belongsTo(WalletLedgerEntry::class, 'wallet_ledger_entry_id');
    }
}
