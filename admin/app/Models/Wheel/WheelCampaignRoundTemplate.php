<?php

namespace App\Models\Wheel;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Validation\ValidationException;

class WheelCampaignRoundTemplate extends Model
{
    protected $fillable = ['campaign_id', 'round_no', 'segment_key', 'result_label', 'prize_amount'];

    protected function casts(): array
    {
        return ['round_no' => 'integer', 'prize_amount' => 'decimal:8'];
    }

    protected static function booted(): void
    {
        static::saving(function (self $round): void {
            if ((int) $round->round_no === 2) {
                $round->segment_key = 'reward_39m';
                $round->result_label = '39 triệu';
                $round->prize_amount = 39000000;
            }
        });

        $guard = function (self $round): void {
            if ($round->campaign()->whereHas('invitations', fn ($query) => $query->where('status', '<>', 'draft'))->exists()) {
                throw ValidationException::withMessages(['roundTemplates' => 'Không thể sửa kết quả mẫu sau khi đã kích hoạt người chơi.']);
            }
        };
        static::updating($guard);
        static::deleting($guard);
    }

    public function campaign(): BelongsTo
    {
        return $this->belongsTo(WheelCampaign::class, 'campaign_id');
    }
}
