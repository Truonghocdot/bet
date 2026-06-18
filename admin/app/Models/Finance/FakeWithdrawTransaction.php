<?php

namespace App\Models\Finance;

use Illuminate\Database\Eloquent\Model;

class FakeWithdrawTransaction extends Model
{
    protected $table = 'fake_withdraw_transactions';

    protected $fillable = [
        'source_index',
        'masked_code',
        'masked_phone',
        'status_label',
        'meta',
    ];

    protected function casts(): array
    {
        return [
            'meta' => 'array',
        ];
    }
}
