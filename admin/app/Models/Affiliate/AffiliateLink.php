<?php

namespace App\Models\Affiliate;

use App\Enum\Affiliate\AffiliateLinkStatus;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\Relations\HasMany;

class AffiliateLink extends Model
{
    protected $fillable = [
        'affiliate_profile_id',
        'campaign_name',
        'tracking_code',
        'landing_url',
        'status',
    ];

    protected function casts(): array
    {
        return [
            'status' => AffiliateLinkStatus::class,
        ];
    }

    public function affiliateProfile(): BelongsTo
    {
        return $this->belongsTo(AffiliateProfile::class);
    }

    public function referrals(): HasMany
    {
        return $this->hasMany(AffiliateReferral::class);
    }
}
