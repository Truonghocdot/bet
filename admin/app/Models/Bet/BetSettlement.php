<?php

namespace App\Models\Bet;

use App\Enum\Bet\BetStatus;
use App\Enum\Bet\SettlementType;
use App\Models\User;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

class BetSettlement extends Model
{
    public $timestamps = false;

    protected $fillable = [
        'ticket_id',
        'period_id',
        'settlement_type',
        'before_status',
        'after_status',
        'payout_amount',
        'profit_loss',
        'note',
        'settled_by',
        'created_at',
    ];

    protected function casts(): array
    {
        return [
            'settlement_type' => SettlementType::class,
            'before_status' => BetStatus::class,
            'after_status' => BetStatus::class,
            'payout_amount' => 'decimal:8',
            'profit_loss' => 'decimal:8',
            'created_at' => 'datetime',
        ];
    }

    public function ticket(): BelongsTo
    {
        return $this->belongsTo(BetTicket::class, 'ticket_id');
    }

    public function period(): BelongsTo
    {
        return $this->belongsTo(GamePeriod::class, 'period_id');
    }

    public function settledBy(): BelongsTo
    {
        return $this->belongsTo(User::class, 'settled_by');
    }
}
