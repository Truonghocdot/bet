<?php

namespace App\Models;

use App\Enum\User\RoleUser;
use App\Enum\User\UserStatus;
use Database\Factories\UserFactory;
use Illuminate\Database\Eloquent\Relations\HasMany;
use Illuminate\Database\Eloquent\Relations\HasOne;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Foundation\Auth\User as Authenticatable;
use Illuminate\Database\Eloquent\SoftDeletes;
use Illuminate\Notifications\Notifiable;

class User extends Authenticatable
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

    public function wallets(): HasMany
    {
        return $this->hasMany(\App\Models\Wallet\Wallet::class);
    }

    public function transactions(): HasMany
    {
        return $this->hasMany(\App\Models\Transaction\Transaction::class);
    }

    public function withdrawalRequests(): HasMany
    {
        return $this->hasMany(\App\Models\Transaction\WithdrawalRequest::class);
    }

    public function gameTickets(): HasMany
    {
        return $this->hasMany(\App\Models\Bet\BetTicket::class);
    }

    public function affiliateProfile(): HasOne
    {
        return $this->hasOne(\App\Models\Affiliate\AffiliateProfile::class);
    }

    public function referredByReferrals(): HasMany
    {
        return $this->hasMany(\App\Models\Affiliate\AffiliateReferral::class, 'referred_user_id');
    }

    public function referralLogs(): HasMany
    {
        return $this->hasMany(\App\Models\Affiliate\AffiliateRewardLog::class, 'referrer_user_id');
    }
}
