<?php

namespace App\Models;

use App\Enum\User\RoleUser;
use App\Enum\User\UserStatus;
use App\Models\Affiliate\AffiliateProfile;
use App\Models\Bet\BetTicket;
use App\Models\Transaction\AccountWithdrawalInfo;
use App\Models\Transaction\Transaction;
use App\Models\Transaction\WithdrawalRequest;
use App\Models\Wallet\Wallet;
use Database\Factories\UserFactory;
use Filament\Models\Contracts\FilamentUser;
use Filament\Panel;
use Illuminate\Database\Eloquent\Relations\HasMany;
use Illuminate\Database\Eloquent\Relations\HasManyThrough;
use Illuminate\Database\Eloquent\Relations\HasOne;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Foundation\Auth\User as Authenticatable;
use Illuminate\Database\Eloquent\SoftDeletes;
use Illuminate\Notifications\Notifiable;
use RuntimeException;

class User extends Authenticatable implements FilamentUser
{
    /** @use HasFactory<UserFactory> */
    use HasFactory, Notifiable, SoftDeletes;

    protected $fillable = [
        'name',
        'email',
        'phone',
        'password',
        'role',
        'status',
    ];

    protected $hidden = [
        'password',
        'remember_token',
    ];

    protected static function booted(): void
    {
        static::creating(function (self $user): void {
            if (filled($user->getKey())) {
                return;
            }

            $user->setAttribute($user->getKeyName(), static::generateUniqueSixDigitId());
        });
    }

    protected function casts(): array
    {
        return [
            'email_verified_at' => 'datetime',
            'phone_verified_at' => 'datetime',
            'last_login_at' => 'datetime',
            'password' => 'hashed',
            'role' => RoleUser::class,
            'status' => UserStatus::class,
        ];
    }

    public function canAccessPanel(Panel $panel): bool
    {
        return $this->deleted_at === null
            && $this->status === UserStatus::ACTIVE
            && in_array($this->role, [
                RoleUser::SUPER_ADMIN,
                RoleUser::ADMIN,
                RoleUser::STAFF,
            ], true);
    }

    public function wallets(): HasMany
    {
        return $this->hasMany(Wallet::class);
    }

    public function transactions(): HasMany
    {
        return $this->hasMany(Transaction::class);
    }

    public function withdrawalRequests(): HasMany
    {
        return $this->hasMany(WithdrawalRequest::class);
    }

    public function gameTickets(): HasMany
    {
        return $this->hasMany(BetTicket::class);
    }

    public function affiliateProfile(): HasOne
    {
        return $this->hasOne(AffiliateProfile::class);
    }

    public function accountWithdrawalInfos(): HasMany
    {
        return $this->hasMany(AccountWithdrawalInfo::class);
    }

    public function affiliateLinks(): HasManyThrough
    {
        return $this->hasManyThrough(
            \App\Models\Affiliate\AffiliateLink::class,
            AffiliateProfile::class,
            'user_id',
            'affiliate_profile_id',
            'id',
            'id',
        );
    }

    public function affiliateReferrals(): HasManyThrough
    {
        return $this->hasManyThrough(
            \App\Models\Affiliate\AffiliateReferral::class,
            AffiliateProfile::class,
            'user_id',
            'affiliate_profile_id',
            'id',
            'id',
        );
    }

    public function affiliateRewardLogs(): HasManyThrough
    {
        return $this->hasManyThrough(
            \App\Models\Affiliate\AffiliateRewardLog::class,
            AffiliateProfile::class,
            'user_id',
            'affiliate_profile_id',
            'id',
            'id',
        );
    }

    public function referredByReferrals(): HasMany
    {
        return $this->hasMany(\App\Models\Affiliate\AffiliateReferral::class, 'referred_user_id');
    }

    public function referredByReferral(): HasOne
    {
        return $this->hasOne(\App\Models\Affiliate\AffiliateReferral::class, 'referred_user_id');
    }

    public function referralLogs(): HasMany
    {
        return $this->hasMany(\App\Models\Affiliate\AffiliateRewardLog::class, 'referrer_user_id');
    }

    public function resolveDirectReferrerUser(): ?self
    {
        $this->loadMissing('referredByReferral.referrerUser');

        $referrer = $this->referredByReferral?->referrerUser;

        return $referrer instanceof self ? $referrer : null;
    }

    public function resolveAgencyOwnerUser(int $maxDepth = 10): ?self
    {
        $currentUser = $this;
        $visitedUserIds = [$this->getKey() => true];

        for ($depth = 0; $depth < $maxDepth; $depth++) {
            $referrer = $currentUser->resolveDirectReferrerUser();

            if (! $referrer) {
                return null;
            }

            if ($referrer->role === RoleUser::AGENCY) {
                return $referrer;
            }

            $referrerId = $referrer->getKey();

            if (isset($visitedUserIds[$referrerId])) {
                return null;
            }

            $visitedUserIds[$referrerId] = true;
            $currentUser = $referrer;
        }

        return null;
    }

    public static function generateUniqueSixDigitId(?int $exceptId = null): int
    {
        for ($attempt = 0; $attempt < 50; $attempt++) {
            $candidate = random_int(100000, 999999);

            $exists = static::query()
                ->when($exceptId !== null, fn ($query) => $query->whereKeyNot($exceptId))
                ->whereKey($candidate)
                ->exists();

            if (! $exists) {
                return $candidate;
            }
        }

        throw new RuntimeException('Không thể tạo user ID 6 số duy nhất.');
    }
}
