<?php

namespace App\Models\Bet;

use App\Enum\Bet\DrawSource;
use App\Enum\Bet\GameType;
use App\Enum\Bet\PeriodStatus;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\HasMany;
use Illuminate\Validation\ValidationException;

class GamePeriod extends Model
{
    protected $fillable = [
        'game_type',
        'period_no',
        'room_code',
        'open_at',
        'close_at',
        'bet_lock_at',
        'draw_at',
        'settled_at',
        'status',
        'draw_source',
        'result_payload',
        'result_hash',
        'manual_result',
    ];

    protected function casts(): array
    {
        return [
            'game_type' => GameType::class,
            'open_at' => 'datetime',
            'close_at' => 'datetime',
            'bet_lock_at' => 'datetime',
            'draw_at' => 'datetime',
            'settled_at' => 'datetime',
            'status' => PeriodStatus::class,
            'draw_source' => DrawSource::class,
            'result_payload' => 'array',
            'manual_result' => 'array',
        ];
    }

    public function tickets(): HasMany
    {
        return $this->hasMany(BetTicket::class, 'period_id');
    }

    public function room(): BelongsTo
    {
        return $this->belongsTo(GameRoom::class, 'room_code', 'code');
    }

    public function items(): HasMany
    {
        return $this->hasMany(BetItem::class, 'period_id');
    }

    public function settlements(): HasMany
    {
        return $this->hasMany(BetSettlement::class, 'period_id');
    }

    public function isAdminMutationLocked(): bool
    {
        $statusValue = $this->status instanceof PeriodStatus ? $this->status->value : (int) $this->status;
        if ($statusValue >= PeriodStatus::LOCKED->value) {
            return true;
        }

        if ($this->bet_lock_at === null) {
            return false;
        }

        return now()->greaterThanOrEqualTo($this->bet_lock_at);
    }

    protected static function booted(): void
    {
        static::updating(function (GamePeriod $period): void {
            if ($period->isAdminMutationLocked()) {
                throw ValidationException::withMessages([
                    'period' => ['Kỳ đã bước vào giai đoạn khóa cược, không thể chỉnh sửa.'],
                ]);
            }
        });

        static::deleting(function (GamePeriod $period): void {
            if ($period->isAdminMutationLocked()) {
                throw ValidationException::withMessages([
                    'period' => ['Kỳ đã bước vào giai đoạn khóa cược, không thể chỉnh sửa.'],
                ]);
            }
        });
    }
}
