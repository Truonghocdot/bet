<?php

namespace App\Filament\Resources\Users\RelationManagers;

use App\Filament\Resources\Wallet\Wallets\WalletResource;
use Filament\Resources\RelationManagers\RelationManager;

class WalletsRelationManager extends RelationManager
{
    protected static string $relationship = 'wallets';
    protected static ?string $relatedResource = WalletResource::class;
    protected static ?string $title = 'Ví';
}
