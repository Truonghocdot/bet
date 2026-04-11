<?php

namespace App\Models\Bet;

use Illuminate\Database\Eloquent\Model;

class GameRoundHistory extends Model
{
    protected $fillable = [
        'game_type',
        'room_code',
        'period_no',
        'result',
        'big_small',
        'color',
        'draw_at',
        'status',
    ];

    protected function casts(): array
    {
        return [
            'draw_at' => 'datetime',
        ];
    }
}
