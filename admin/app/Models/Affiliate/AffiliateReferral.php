<?php

namespace App\Models\Affiliate;

use App\Enum\Affiliate\AffiliateReferralStatus;
use App\Models\Transaction\Transaction;
use App\Models\User;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

class AffiliateReferral extends Model
{
    protected $fillable = [
        'affiliate_profile_id',
        'referrer_user_id',
        'referred_user_id',
        'affiliate_link_id',
        'first_deposit_transaction_id',
        'first_deposit_amount',
        'qualified_at',
        'status',
    ];

    protected function casts(): array
    {
        return [
            'first_deposit_amount' => 'decimal:8',
            'qualified_at' => 'datetime',
            'status' => AffiliateReferralStatus::class,
        ];
    }

    public function affiliateProfile(): BelongsTo
    {
        return $this->belongsTo(AffiliateProfile::class);
    }

    public function referrerUser(): BelongsTo
    {
        return $this->belongsTo(User::class, 'referrer_user_id');
    }

    public function referredUser(): BelongsTo
    {
        return $this->belongsTo(User::class, 'referred_user_id');
    }

    public function affiliateLink(): BelongsTo
    {
        return $this->belongsTo(AffiliateLink::class);
    }

    public function firstDepositTransaction(): BelongsTo
    {
        return $this->belongsTo(Transaction::class, 'first_deposit_transaction_id');
    }
}
