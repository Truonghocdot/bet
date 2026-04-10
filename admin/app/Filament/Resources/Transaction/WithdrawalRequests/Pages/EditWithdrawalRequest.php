<?php

namespace App\Filament\Resources\Transaction\WithdrawalRequests\Pages;

use App\Filament\Resources\Transaction\WithdrawalRequests\WithdrawalRequestResource;
use Filament\Resources\Pages\EditRecord;

class EditWithdrawalRequest extends EditRecord
{
    protected static string $resource = WithdrawalRequestResource::class;
    protected static ?string $title = 'Xử lý yêu cầu rút';
}
