<?php

namespace App\Filament\Resources\Payment\PaymentReceivingAccounts\Pages;

use App\Filament\Resources\Payment\PaymentReceivingAccounts\PaymentReceivingAccountResource;
use Filament\Resources\Pages\CreateRecord;

class CreatePaymentReceivingAccount extends CreateRecord
{
    protected static string $resource = PaymentReceivingAccountResource::class;
}
