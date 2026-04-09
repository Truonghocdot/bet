<?php

namespace App\Models\Affiliate;

use App\Enum\Affiliate\AffiliateProfileStatus;
use App\Models\User;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\Relations\HasMany;

class AffiliateProfile extends Model
{
    protected $fillable = [
        'user_id',
        'ref_code',
        'ref_link',
        'status',
        'approved_by',
        'approved_at',
    ];

    protected function casts(): array
    {
        return [
            'status' => AffiliateProfileStatus::class,
            'approved_at' => 'datetime',
        ];
    }

    public function user(): BelongsTo
    {
        return $this->belongsTo(User::class);
    }

    public function approvedBy(): BelongsTo
    {
        return $this->belongsTo(User::class, 'approved_by');
    }

    public function links(): HasMany
    {
        return $this->hasMany(AffiliateLink::class);
    }

    public function referrals(): HasMany
    {
        return $this->hasMany(AffiliateReferral::class);
    }

    public function rewardLogs(): HasMany
    {
        return $this->hasMany(AffiliateRewardLog::class);
    }
}
