<?php

namespace App\Filament\Resources\Transaction\WithdrawalRequests\Pages;

use App\Filament\Resources\Transaction\WithdrawalRequests\WithdrawalRequestResource;
use Filament\Resources\Pages\ListRecords;

class ListWithdrawalRequests extends ListRecords
{
    protected static string $resource = WithdrawalRequestResource::class;
}
