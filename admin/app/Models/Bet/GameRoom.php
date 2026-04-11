<?php

namespace App\Models\Bet;

use App\Enum\Bet\GameType;
use App\Enum\Bet\RoomStatus;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\HasMany;

class GameRoom extends Model
{
    protected $fillable = [
        'code',
        'game_type',
        'duration_seconds',
        'bet_cutoff_seconds',
        'status',
        'sort_order',
    ];

    protected function casts(): array
    {
        return [
            'game_type' => GameType::class,
            'status' => RoomStatus::class,
        ];
    }

    public function periods(): HasMany
    {
        return $this->hasMany(GamePeriod::class, 'room_code', 'code');
    }
}
