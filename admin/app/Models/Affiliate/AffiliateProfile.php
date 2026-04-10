<?php

namespace App\Models\Affiliate;

use App\Enum\Affiliate\AffiliateProfileStatus;
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
    ];

    protected function casts(): array
    {
        return [
            'status' => AffiliateProfileStatus::class,
        ];
    }

    protected static function booted(): void
    {
        static::creating(function (self $profile): void {
            $identity = static::generateReferralIdentity($profile->ref_code, $profile->ref_link);

            $profile->ref_code = $identity['ref_code'];
            $profile->ref_link = $identity['ref_link'];
        });
    }

    public function user(): BelongsTo
    {
        return $this->belongsTo(User::class);
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

    public static function generateReferralIdentity(?string $refCode = null, ?string $refLink = null): array
    {
        $resolvedCode = filled($refCode) ? trim((string) $refCode) : static::generateUniqueRefCode();
        $resolvedLink = filled($refLink)
            ? trim((string) $refLink)
            : rtrim((string) config('app.url'), '/').'/register?ref='.$resolvedCode;

        return [
            'ref_code' => $resolvedCode,
            'ref_link' => $resolvedLink,
        ];
    }

    public static function generateUniqueRefCode(): string
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
