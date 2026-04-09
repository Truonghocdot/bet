<?php

namespace App\Filament\Resources\Transaction\WithdrawalRequests\Pages;

use App\Filament\Resources\Transaction\WithdrawalRequests\WithdrawalRequestResource;
use Filament\Resources\Pages\CreateRecord;

class CreateWithdrawalRequest extends CreateRecord
{
    protected static string $resource = WithdrawalRequestResource::class;
}
