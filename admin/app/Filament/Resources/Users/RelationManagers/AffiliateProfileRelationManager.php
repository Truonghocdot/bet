<?php

namespace App\Filament\Resources\Users\RelationManagers;

use App\Enum\Affiliate\AffiliateProfileStatus;
use App\Support\Filament\EnumPresenter;
use Filament\Actions\CreateAction;
use Filament\Actions\EditAction;
use Filament\Forms\Components\Hidden;
use Filament\Schemas\Components\Section;
use Filament\Forms\Components\Select;
use Filament\Resources\RelationManagers\RelationManager;
use Filament\Schemas\Schema;
use Filament\Tables\Columns\TextColumn;
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
            ->columns([
                TextColumn::make('ref_code')
                    ->label('Mã giới thiệu')
                    ->searchable()
                    ->copyable()
                    ->fontFamily('mono'),
                TextColumn::make('ref_link')
                    ->label('Link giới thiệu')
                    ->copyable()
                    ->wrap(),
                TextColumn::make('status')
                    ->label('Trạng thái')
                    ->badge()
                    ->formatStateUsing(fn ($state): string => EnumPresenter::label(AffiliateProfileStatus::class, $state))
                    ->color(fn ($state): string => EnumPresenter::color(AffiliateProfileStatus::class, $state)),
                TextColumn::make('referrals_count')
                    ->label('Người chơi trực thuộc')
                    ->counts('referrals'),
            ])
            ->headerActions([
                CreateAction::make()
                    ->visible(fn (): bool => blank($this->getOwnerRecord()->affiliateProfile)),
            ])
            ->recordActions([
                EditAction::make(),
            ])
            ->paginated(false)
            ->emptyStateHeading('Chưa có hồ sơ affiliate')
            ->emptyStateDescription('Tạo hồ sơ affiliate để bắt đầu quản lý tuyến người chơi của đại lý này.');
    }
}
