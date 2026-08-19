<?php

namespace App\Models\Wheel;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Validation\ValidationException;

class WheelInvitationRound extends Model
{
    protected $fillable = ['invitation_id', 'round_no', 'segment_key', 'result_label', 'prize_amount', 'status', 'spun_at'];

    protected function casts(): array
    {
        return ['round_no' => 'integer', 'prize_amount' => 'decimal:8', 'spun_at' => 'datetime'];
    }

    protected static function booted(): void
    {
        $guard = function (self $round): void {
            if ($round->invitation()->where('status', '<>', 'draft')->exists()) {
                throw ValidationException::withMessages(['rounds' => 'Kết quả đã bị khóa từ lúc kích hoạt lời mời.']);
            }
        };
        static::updating($guard);
        static::deleting($guard);
    }

    public function invitation(): BelongsTo
    {
        return $this->belongsTo(WheelInvitation::class, 'invitation_id');
    }
}
