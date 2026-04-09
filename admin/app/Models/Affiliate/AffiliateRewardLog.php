<?php

namespace App\Models\Affiliate;

use App\Enum\Affiliate\AffiliateRewardStatus;
use App\Enum\Wallet\UnitTransaction;
use App\Models\User;
use App\Models\Wallet\WalletLedgerEntry;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

class AffiliateRewardLog extends Model
{
    protected $fillable = [
        'affiliate_profile_id',
        'referrer_user_id',
        'setting_id',
        'required_qualified_referrals',
        'actual_qualified_referrals',
        'reward_amount',
        'unit',
        'status',
        'wallet_ledger_entry_id',
        'granted_at',
    ];

    protected function casts(): array
    {
        return [
            'required_qualified_referrals' => 'integer',
            'actual_qualified_referrals' => 'integer',
            'reward_amount' => 'decimal:8',
            'unit' => UnitTransaction::class,
            'status' => AffiliateRewardStatus::class,
            'granted_at' => 'datetime',
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

    public function setting(): BelongsTo
    {
        return $this->belongsTo(AffiliateRewardSetting::class, 'setting_id');
    }

    public function walletLedgerEntry(): BelongsTo
    {
        return $this->belongsTo(WalletLedgerEntry::class);
    }
}
