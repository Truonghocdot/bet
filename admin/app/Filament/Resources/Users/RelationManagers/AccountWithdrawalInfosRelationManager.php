<?php

namespace App\Filament\Resources\Users\RelationManagers;

use App\Enum\Wallet\UnitTransaction;
use App\Filament\Resources\Transaction\AccountWithdrawalInfos\AccountWithdrawalInfoResource;
use App\Support\Filament\EnumPresenter;
use Filament\Forms\Components\Hidden;
use Filament\Schemas\Components\Section;
use Filament\Forms\Components\Select;
use Filament\Forms\Components\TextInput;
use Filament\Forms\Components\Toggle;
use Filament\Resources\RelationManagers\RelationManager;
use Filament\Schemas\Schema;
use Filament\Tables\Table;

class AccountWithdrawalInfosRelationManager extends RelationManager
{
    protected static string $relationship = 'accountWithdrawalInfos';
    protected static ?string $relatedResource = AccountWithdrawalInfoResource::class;
    protected static ?string $title = 'Tài khoản rút';

    public function form(Schema $schema): Schema
    {
        return $schema->components([
            Section::make('Thông tin tài khoản rút')
                ->schema([
                    Hidden::make('user_id')
                        ->default(fn ($livewire) => $livewire->getOwnerRecord()->getKey()),
                    Select::make('unit')
                        ->label('Đơn vị')
                        ->options(EnumPresenter::options(UnitTransaction::class))
                        ->required(),
                    TextInput::make('provider_code')
                        ->label('Mã nhà cung cấp')
                        ->maxLength(50)
                        ->required(),
                    TextInput::make('account_name')
                        ->label('Chủ tài khoản')
                        ->maxLength(255)
                        ->required(),
                    TextInput::make('account_number')
                        ->label('Số tài khoản')
                        ->maxLength(255)
                        ->required(),
                    Toggle::make('is_default')
                        ->label('Mặc định'),
                ])
                ->columns(2),
        ]);
    }

    public function table(Table $table): Table
    {
        return $table->headerActions([
            \Filament\Actions\CreateAction::make(),
        ]);
    }
}
