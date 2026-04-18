<?php

namespace App\Models\Bet;

use App\Enum\Bet\GameType;
use App\Enum\Bet\RoomStatus;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\HasMany;

class GameRoom extends Model
{
    public const CONTROL_LOCK_REDIS_PREFIX = 'admin:lock:control_room:';

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

    public function controlLabel(): string
    {
        $code = strtolower(trim((string) $this->code));
        $match = [];

        if (preg_match('/^(wingo|k3|lottery)_(\d+)(s|m)$/', $code, $match) === 1) {
            $game = match ($match[1] ?? '') {
                'wingo' => 'Wingo',
                'k3' => 'K3',
                'lottery' => '5D',
                default => strtoupper((string) ($match[1] ?? 'ROOM')),
            };

            $duration = (string) ($match[2] ?? '');
            $unit = strtolower((string) ($match[3] ?? ''));
            $suffix = $unit === 's' ? 's' : 'm';

            return trim(sprintf('%s %s%s', $game, $duration, $suffix));
        }

        return strtoupper($this->code);
    }

    public static function controlLockRedisKey(string $roomCode): string
    {
        return self::CONTROL_LOCK_REDIS_PREFIX.trim($roomCode);
    }
}
