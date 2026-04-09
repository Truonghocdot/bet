<?php

namespace App\Filament\Resources\Users\RelationManagers;

use App\Filament\Resources\Transaction\Transactions\TransactionResource;
use Filament\Resources\RelationManagers\RelationManager;

class TransactionsRelationManager extends RelationManager
{
    protected static string $relationship = 'transactions';
    protected static ?string $relatedResource = TransactionResource::class;
    protected static ?string $title = 'Giao dịch';
}
