<?php

namespace App\Models\Bet;

use App\Enum\Bet\BetStatus;
use App\Enum\Bet\BetTicketType;
use App\Enum\Bet\GameType;
use App\Enum\Wallet\UnitTransaction;
use App\Models\User;
use App\Models\Wallet\Wallet;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\Relations\HasMany;

class BetTicket extends Model
{
    protected $fillable = [
        'ticket_no',
        'user_id',
        'wallet_id',
        'unit',
        'game_type',
        'period_id',
        'bet_type',
        'stake',
        'total_odds',
        'potential_payout',
        'actual_payout',
        'status',
        'placed_ip',
        'placed_device',
        'settled_at',
    ];

    protected function casts(): array
    {
        return [
            'game_type' => GameType::class,
            'unit' => UnitTransaction::class,
            'bet_type' => BetTicketType::class,
            'stake' => 'decimal:8',
            'total_odds' => 'decimal:6',
            'potential_payout' => 'decimal:8',
            'actual_payout' => 'decimal:8',
            'status' => BetStatus::class,
            'settled_at' => 'datetime',
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

    public function period(): BelongsTo
    {
        return $this->belongsTo(GamePeriod::class, 'period_id');
    }

    public function items(): HasMany
    {
        return $this->hasMany(BetItem::class, 'ticket_id');
    }

    public function settlements(): HasMany
    {
        return $this->hasMany(BetSettlement::class, 'ticket_id');
    }
}
