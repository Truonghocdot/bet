<?php

namespace App\Filament\Resources\Users\RelationManagers;

use App\Enum\Affiliate\AffiliateProfileStatus;
use App\Support\Filament\EnumPresenter;
use Filament\Actions\CreateAction;
use Filament\Forms\Components\Hidden;
use Filament\Schemas\Components\Section;
use Filament\Forms\Components\Select;
use Filament\Resources\RelationManagers\RelationManager;
use Filament\Schemas\Schema;
use Filament\Tables\Table;

class AffiliateProfileRelationManager extends RelationManager
{
    protected static string $relationship = 'affiliateProfile';
    protected static ?string $title = 'Hồ sơ affiliate';

    public function form(Schema $schema): Schema
    {
        return $schema->components([
            Section::make('Hồ sơ affiliate')
                ->schema([
                    Hidden::make('user_id')
                        ->default(fn ($livewire) => $livewire->getOwnerRecord()->getKey()),
                    Select::make('status')
                        ->label('Trạng thái')
                        ->options(EnumPresenter::options(AffiliateProfileStatus::class))
                        ->required(),
                ])
                ->columns(1),
        ]);
    }

    public function table(Table $table): Table
    {
        return $table
            ->headerActions([
                CreateAction::make()
                    ->visible(fn (): bool => blank($this->getOwnerRecord()->affiliateProfile)),
            ]);
    }
}
