<?php

namespace App\Filament\Resources\Users\RelationManagers;

use App\Filament\Resources\Transaction\WithdrawalRequests\WithdrawalRequestResource;
use Filament\Resources\RelationManagers\RelationManager;

class WithdrawalRequestsRelationManager extends RelationManager
{
    protected static string $relationship = 'withdrawalRequests';
    protected static ?string $relatedResource = WithdrawalRequestResource::class;
    protected static ?string $title = 'Yêu cầu rút';
}
