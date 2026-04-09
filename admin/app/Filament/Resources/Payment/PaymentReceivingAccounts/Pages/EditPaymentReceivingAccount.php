<?php

namespace App\Filament\Resources\Payment\PaymentReceivingAccounts\Pages;

use App\Filament\Resources\Payment\PaymentReceivingAccounts\PaymentReceivingAccountResource;
use Filament\Actions\DeleteAction;
use Filament\Actions\ForceDeleteAction;
use Filament\Actions\RestoreAction;
use Filament\Resources\Pages\EditRecord;

class EditPaymentReceivingAccount extends EditRecord
{
    protected static string $resource = PaymentReceivingAccountResource::class;

    protected function getHeaderActions(): array
    {
        return [
            DeleteAction::make(),
            ForceDeleteAction::make(),
            RestoreAction::make(),
        ];
    }
}
