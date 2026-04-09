<?php

namespace App\Filament\Resources\Payment\PaymentReceivingAccounts\Pages;

use App\Filament\Resources\Payment\PaymentReceivingAccounts\PaymentReceivingAccountResource;
use Filament\Actions\CreateAction;
use Filament\Resources\Pages\ListRecords;

class ListPaymentReceivingAccounts extends ListRecords
{
    protected static string $resource = PaymentReceivingAccountResource::class;

    protected function getHeaderActions(): array
    {
        return [
            CreateAction::make(),
        ];
    }
}
