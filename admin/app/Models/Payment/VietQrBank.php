<?php

namespace App\Models\Payment;

use Illuminate\Database\Eloquent\Factories\HasFactory;
use Illuminate\Database\Eloquent\Model;

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
}
