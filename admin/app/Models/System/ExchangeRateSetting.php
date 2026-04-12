<?php

namespace App\Models\System;

use App\Models\User;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

class ExchangeRateSetting extends Model
{
    use HasFactory;

    public const CODE = 'USDT_VND';

    protected $table = 'exchange_rate_settings';

    protected $fillable = [
        'code',
        'base_currency',
        'quote_currency',
        'rate',
        'source_rate',
        'auto_sync',
        'source_name',
        'last_synced_at',
        'updated_by',
        'note',
        'nowpayments_api_key',
        'nowpayments_ipn_secret',
        'nowpayments_payout_wallet',
        'nowpayments_sandbox',
    ];

    protected $casts = [
        'rate' => 'decimal:8',
        'source_rate' => 'decimal:8',
        'auto_sync' => 'boolean',
        'last_synced_at' => 'datetime',
        'nowpayments_sandbox' => 'boolean',
    ];

    public function updatedBy(): BelongsTo
    {
        return $this->belongsTo(User::class, 'updated_by');
    }
}
