<?php

namespace App\Models\Affiliate;

use App\Enum\Wallet\UnitTransaction;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\HasMany;

class AffiliateRewardSetting extends Model
{
    protected $fillable = [
        'name',
        'required_qualified_referrals',
        'reward_amount',
        'unit',
        'is_active',
        'effective_from',
        'effective_to',
        'note',
    ];

    protected function casts(): array
    {
        return [
            'reward_amount' => 'decimal:8',
            'unit' => UnitTransaction::class,
            'is_active' => 'boolean',
            'effective_from' => 'datetime',
            'effective_to' => 'datetime',
        ];
    }

    public function rewardLogs(): HasMany
    {
        return $this->hasMany(AffiliateRewardLog::class, 'setting_id');
    }
}
