<?php

namespace App\Filament\Resources\Transaction\AccountWithdrawalInfos\Pages;

use App\Filament\Resources\Transaction\AccountWithdrawalInfos\AccountWithdrawalInfoResource;
use Filament\Actions\CreateAction;
use Filament\Resources\Pages\ListRecords;

class ListAccountWithdrawalInfos extends ListRecords
{
    protected static string $resource = AccountWithdrawalInfoResource::class;

    protected function getHeaderActions(): array
    {
        return [
            CreateAction::make(),
        ];
    }
}
