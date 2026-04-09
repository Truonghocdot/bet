<?php

namespace App\Models\Bet;

use App\Enum\Bet\DrawSource;
use App\Enum\Bet\GameType;
use App\Enum\Bet\PeriodStatus;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\HasMany;

class GamePeriod extends Model
{
    protected $fillable = [
        'game_type',
        'period_no',
        'room_code',
        'open_at',
        'close_at',
        'draw_at',
        'settled_at',
        'status',
        'draw_source',
        'result_payload',
        'result_hash',
    ];

    protected function casts(): array
    {
        return [
            'game_type' => GameType::class,
            'open_at' => 'datetime',
            'close_at' => 'datetime',
            'draw_at' => 'datetime',
            'settled_at' => 'datetime',
            'status' => PeriodStatus::class,
            'draw_source' => DrawSource::class,
            'result_payload' => 'array',
        ];
    }

    public function tickets(): HasMany
    {
        return $this->hasMany(BetTicket::class, 'period_id');
    }

    public function items(): HasMany
    {
        return $this->hasMany(BetItem::class, 'period_id');
    }

    public function settlements(): HasMany
    {
        return $this->hasMany(BetSettlement::class, 'period_id');
    }
}
