<?php

namespace App\Filament\Resources\Users\RelationManagers;

use App\Enum\Affiliate\AffiliateProfileStatus;
use App\Filament\Resources\Affiliate\AffiliateProfiles\AffiliateProfileResource;
use App\Support\Filament\EnumPresenter;
use Filament\Actions\CreateAction;
use Filament\Forms\Components\DateTimePicker;
use Filament\Forms\Components\Hidden;
use Filament\Schemas\Components\Section;
use Filament\Forms\Components\Select;
use Filament\Forms\Components\TextInput;
use Filament\Resources\RelationManagers\RelationManager;
use Filament\Schemas\Schema;
use Filament\Tables\Table;

class AffiliateProfileRelationManager extends RelationManager
{
    protected static string $relationship = 'affiliateProfile';
    protected static ?string $relatedResource = AffiliateProfileResource::class;
    protected static ?string $title = 'Hồ sơ affiliate';

    public function form(Schema $schema): Schema
    {
        return $schema->components([
            Section::make('Hồ sơ affiliate')
                ->schema([
                    Hidden::make('user_id')
                        ->default(fn ($livewire) => $livewire->getOwnerRecord()->getKey()),
                    TextInput::make('ref_code')
                        ->label('Mã giới thiệu')
                        ->required()
                        ->maxLength(50),
                    TextInput::make('ref_link')
                        ->label('Link giới thiệu')
                        ->required()
                        ->maxLength(255),
                    Select::make('status')
                        ->label('Trạng thái')
                        ->options(EnumPresenter::options(AffiliateProfileStatus::class))
                        ->required(),
                    Select::make('approved_by')
                        ->label('Duyệt bởi')
                        ->relationship('approvedBy', 'name')
                        ->searchable()
                        ->preload(),
                    DateTimePicker::make('approved_at')
                        ->label('Duyệt lúc'),
                ])
                ->columns(2),
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
