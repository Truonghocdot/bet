<?php

namespace App\Models\Bet;

use App\Enum\Bet\BetItemResult;
use App\Enum\Bet\BetOptionType;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

class BetItem extends Model
{
    protected $fillable = [
        'ticket_id',
        'period_id',
        'option_type',
        'option_key',
        'option_label',
        'odds_at_placement',
        'stake',
        'result',
        'payout_amount',
        'result_payload',
        'settled_at',
    ];

    protected function casts(): array
    {
        return [
            'option_type' => BetOptionType::class,
            'odds_at_placement' => 'decimal:4',
            'stake' => 'decimal:8',
            'result' => BetItemResult::class,
            'payout_amount' => 'decimal:8',
            'result_payload' => 'array',
            'settled_at' => 'datetime',
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
}
