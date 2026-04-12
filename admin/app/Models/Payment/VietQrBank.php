<?php

namespace App\Models\Payment;

use App\Enum\Payment\PaymentReceivingAccountType;
use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\HasMany;

class VietQrBank extends Model
{
    use HasFactory;

    protected $table = 'vietqr_banks';

    protected $fillable = [
        'source_id',
        'code',
        'name',
        'short_name',
        'bin',
        'logo',
        'transfer_supported',
        'lookup_supported',
        'support',
        'raw_payload',
        'synced_at',
    ];

    protected function casts(): array
    {
        return [
            'source_id' => 'integer',
            'transfer_supported' => 'boolean',
            'lookup_supported' => 'boolean',
            'support' => 'integer',
            'raw_payload' => 'array',
            'synced_at' => 'datetime',
        ];
    }

    public function receivingAccounts(): HasMany
    {
        return $this->hasMany(PaymentReceivingAccount::class, 'provider_code', 'code')
            ->where('type', PaymentReceivingAccountType::BANK->value);
    }
}
