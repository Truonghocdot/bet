<?php

namespace App\Filament\Resources\Users\Tables;

use App\Enum\User\RoleUser;
use App\Models\User;
use App\Enum\User\UserStatus;
use App\Support\Filament\EnumPresenter;
use Filament\Actions\BulkActionGroup;
use Filament\Actions\DeleteBulkAction;
use Filament\Actions\EditAction;
use Filament\Actions\ForceDeleteBulkAction;
use Filament\Actions\RestoreBulkAction;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Filters\TrashedFilter;
use Filament\Tables\Table;
use Illuminate\Database\Eloquent\Builder;

class UsersTable
{
    private const REFERRAL_SEARCH_DEPTH = 8;

    public static function configure(Table $table, ?RoleUser $fixedRole = null): Table
    {
        return $table
            ->modifyQueryUsing(function (Builder $query) use ($fixedRole): Builder {
                if ($fixedRole !== RoleUser::CLIENT) {
                    return $query;
                }

                return $query->with(self::referralRelationChain());
            })
            ->columns([
                TextColumn::make('id')
                    ->label('ID')
                    ->sortable(),
                TextColumn::make('name')
                    ->label('Họ và tên')
                    ->searchable()
                    ->sortable(),
                TextColumn::make('phone')
                    ->label('Số điện thoại')
                    ->searchable()
                    ->copyable()
                    ->toggleable(),
                TextColumn::make('agency_name')
                    ->label('Thuộc đại lý')
                    ->visible($fixedRole === RoleUser::CLIENT)
                    ->getStateUsing(function (User $record): string {
                        $agency = $record->resolveAgencyOwnerUser();

                        if (! $agency) {
                            return '—';
                        }

                        return ($agency->name ?: 'Đại lý').' (#'.$agency->id.')';
                    })
                    ->searchable(query: function (Builder $query, string $search): void {
                        self::applyAgencyOwnerSearch($query, $search);
                    }),
                TextColumn::make('direct_referrer_name')
                    ->label('Người mời trực tiếp')
                    ->visible($fixedRole === RoleUser::CLIENT)
                    ->getStateUsing(function (User $record): string {
                        $referrer = $record->resolveDirectReferrerUser();

                        if (! $referrer) {
                            return '—';
                        }

                        $fallbackName = $referrer->role === RoleUser::AGENCY ? 'Đại lý' : 'Người chơi';

                        return ($referrer->name ?: $fallbackName).' (#'.$referrer->id.')';
                    })
                    ->searchable(query: function (Builder $query, string $search): void {
                        $query->whereHas('referredByReferral.referrerUser', function (Builder $referrerQuery) use ($search): void {
                            $referrerQuery->where(function (Builder $innerQuery) use ($search): void {
                                $innerQuery
                                    ->where('name', 'like', '%'.$search.'%')
                                    ->orWhere('id', 'like', '%'.$search.'%');
                            });
                        });
                    }),
                TextColumn::make('role')
                    ->label('Vai trò')
                    ->badge()
                    ->formatStateUsing(fn ($state): string => EnumPresenter::label(RoleUser::class, $state))
                    ->color(fn ($state): string => EnumPresenter::color(RoleUser::class, $state))
                    ->visible($fixedRole === null),
                TextColumn::make('status')
                    ->label('Trạng thái')
                    ->badge()
                    ->formatStateUsing(fn ($state): string => EnumPresenter::label(UserStatus::class, $state))
                    ->color(fn ($state): string => EnumPresenter::color(UserStatus::class, $state)),
                TextColumn::make('phone_verified_at')
                    ->label('SĐT xác minh')
                    ->dateTime()
                    ->toggleable(),
                TextColumn::make('last_login_at')
                    ->label('Đăng nhập cuối')
                    ->dateTime()
                    ->toggleable(),
                TextColumn::make('wallets_count')
                    ->label('Số ví')
                    ->counts('wallets')
                    ->sortable(),
                TextColumn::make('transactions_count')
                    ->label('Số giao dịch')
                    ->counts('transactions')
                    ->sortable(),
                TextColumn::make('withdrawal_requests_count')
                    ->label('Số lệnh rút')
                    ->counts('withdrawalRequests')
                    ->sortable(),
                TextColumn::make('game_tickets_count')
                    ->label('Số vé cược')
                    ->counts('gameTickets')
                    ->sortable(),
                TextColumn::make('created_at')
                    ->label('Tạo lúc')
                    ->dateTime()
                    ->sortable(),
            ])
            ->filters([
                TrashedFilter::make(),
            ])
            ->defaultSort('id', 'desc')
            ->poll(2000)
            ->recordActions([
                EditAction::make(),
            ])
            ->toolbarActions([
                BulkActionGroup::make([
                    DeleteBulkAction::make(),
                    ForceDeleteBulkAction::make(),
                    RestoreBulkAction::make(),
                ]),
            ]);
    }

    /**
     * @return array<int, string>
     */
    private static function referralRelationChain(): array
    {
        $relations = [];
        $path = 'referredByReferral.referrerUser';

        for ($depth = 0; $depth < self::REFERRAL_SEARCH_DEPTH; $depth++) {
            $relations[] = $path;
            $path .= '.referredByReferral.referrerUser';
        }

        return $relations;
    }

    private static function applyAgencyOwnerSearch(Builder $query, string $search): void
    {
        $relationPath = 'referredByReferral.referrerUser';

        $query->where(function (Builder $outerQuery) use ($search, $relationPath): void {
            for ($depth = 0; $depth < self::REFERRAL_SEARCH_DEPTH; $depth++) {
                $outerQuery->orWhereHas($relationPath, function (Builder $agencyQuery) use ($search): void {
                    $agencyQuery
                        ->where('role', RoleUser::AGENCY->value)
                        ->where(function (Builder $innerQuery) use ($search): void {
                            $innerQuery
                                ->where('name', 'like', '%'.$search.'%')
                                ->orWhere('id', 'like', '%'.$search.'%');
                        });
                });

                $relationPath .= '.referredByReferral.referrerUser';
            }
        });
    }
}
