<?php

namespace App\Models\Affiliate;

use App\Enum\Affiliate\AffiliateProfileStatus;
use App\Models\User;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\Relations\HasMany;
use Illuminate\Support\Str;

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

    protected static function booted(): void
    {
        static::creating(function (self $profile): void {
            if (blank($profile->ref_code)) {
                $profile->ref_code = static::generateUniqueRefCode();
            }

            if (blank($profile->ref_link)) {
                $profile->ref_link = rtrim((string) config('app.url'), '/').'/register?ref='.$profile->ref_code;
            }
        });
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

    private static function generateUniqueRefCode(): string
    {
        for ($attempt = 0; $attempt < 20; $attempt++) {
            $code = 'REF'.Str::upper(Str::random(8));

            if (! static::query()->where('ref_code', $code)->exists()) {
                return $code;
            }
        }

        return 'REF'.Str::upper(Str::random(12));
    }
}
