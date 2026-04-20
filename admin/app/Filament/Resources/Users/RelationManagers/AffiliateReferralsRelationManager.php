<?php

namespace App\Filament\Resources\Users\RelationManagers;

use App\Enum\Affiliate\AffiliateReferralStatus;
use App\Support\Filament\EnumPresenter;
use Filament\Resources\RelationManagers\RelationManager;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Filters\SelectFilter;
use Filament\Tables\Table;
use Illuminate\Database\Eloquent\Builder;
use Illuminate\Database\Eloquent\Model;

class AffiliateReferralsRelationManager extends RelationManager
{
    protected static string $relationship = 'affiliateReferrals';
    protected static ?string $title = 'Người chơi trực thuộc';

    public static function canViewForRecord(Model $ownerRecord, string $pageClass): bool
    {
        return filled($ownerRecord->affiliateProfile);
    }

    public function table(Table $table): Table
    {
        return $table
            ->modifyQueryUsing(fn (Builder $query): Builder => $query->with([
                'referredUser:id,name,phone',
                'firstDepositTransaction:id,client_ref',
            ]))
            ->columns([
                TextColumn::make('referredUser.id')
                    ->label('ID người chơi')
                    ->sortable(),
                TextColumn::make('referredUser.name')
                    ->label('Người chơi')
                    ->searchable(query: function (Builder $query, string $search): void {
                        $query->whereHas('referredUser', function (Builder $userQuery) use ($search): void {
                            $userQuery
                                ->where('name', 'like', '%'.$search.'%')
                                ->orWhere('phone', 'like', '%'.$search.'%');
                        });
                    })
                    ->sortable(),
                TextColumn::make('referredUser.phone')
                    ->label('Số điện thoại')
                    ->toggleable(),
                TextColumn::make('status')
                    ->label('Trạng thái')
                    ->badge()
                    ->formatStateUsing(fn ($state): string => EnumPresenter::label(AffiliateReferralStatus::class, $state))
                    ->color(fn ($state): string => EnumPresenter::color(AffiliateReferralStatus::class, $state)),
                TextColumn::make('first_deposit_amount')
                    ->label('Nạp đầu')
                    ->money('VND')
                    ->toggleable(),
                TextColumn::make('firstDepositTransaction.client_ref')
                    ->label('Mã GD nạp đầu')
                    ->toggleable(),
                TextColumn::make('qualified_at')
                    ->label('Đạt điều kiện lúc')
                    ->dateTime()
                    ->toggleable(),
                TextColumn::make('created_at')
                    ->label('Tham gia lúc')
                    ->dateTime()
                    ->sortable(),
            ])
            ->filters([
                SelectFilter::make('status')
                    ->label('Trạng thái')
                    ->options(EnumPresenter::options(AffiliateReferralStatus::class)),
            ])
            ->defaultSort('created_at', 'desc')
            ->emptyStateHeading('Chưa có người chơi trực thuộc')
            ->emptyStateDescription('Khi người chơi đăng ký hoặc nạp dưới tuyến đại lý này, danh sách sẽ hiển thị tại đây.');
    }
}
