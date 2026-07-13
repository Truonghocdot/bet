<?php

namespace App\Filament\Resources\System\Notifications\RelationManagers;

use App\Enum\Notification\NotificationAudience;
use App\Enum\Notification\NotificationResponseStatus;
use App\Support\Filament\EnumPresenter;
use Filament\Resources\RelationManagers\RelationManager;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Filters\SelectFilter;
use Filament\Tables\Table;
use Illuminate\Database\Eloquent\Builder;
use Illuminate\Database\Eloquent\Model;

class TargetUsersRelationManager extends RelationManager
{
    protected static string $relationship = 'targetUsers';
    protected static ?string $title = 'Người dùng đích & trạng thái phản hồi';

    public static function canViewForRecord(Model $ownerRecord, string $pageClass): bool
    {
        return (int) ($ownerRecord->audience?->value ?? $ownerRecord->audience) === NotificationAudience::USERS->value;
    }

    public function table(Table $table): Table
    {
        return $table
            ->columns([
                TextColumn::make('id')
                    ->label('ID')
                    ->sortable(),
                TextColumn::make('name')
                    ->label('Người dùng')
                    ->searchable(query: function (Builder $query, string $search): void {
                        $query->where(function (Builder $builder) use ($search): void {
                            $builder
                                ->where('users.name', 'like', '%'.$search.'%')
                                ->orWhere('users.phone', 'like', '%'.$search.'%');
                        });
                    })
                    ->sortable(),
                TextColumn::make('phone')
                    ->label('Số điện thoại')
                    ->toggleable(),
                TextColumn::make('pivot.response_status')
                    ->label('Phản hồi')
                    ->badge()
                    ->formatStateUsing(function ($state): string {
                        if (blank($state)) {
                            return '—';
                        }

                        return EnumPresenter::label(NotificationResponseStatus::class, $state);
                    })
                    ->color(function ($state): string {
                        if (blank($state)) {
                            return 'gray';
                        }

                        return EnumPresenter::color(NotificationResponseStatus::class, $state);
                    }),
                TextColumn::make('pivot.responded_at')
                    ->label('Phản hồi lúc')
                    ->dateTime(format: 'd/m/Y H:i', timezone: config('app.timezone', 'Asia/Ho_Chi_Minh'))
                    ->placeholder('—'),
            ])
            ->filters([
                SelectFilter::make('response_status')
                    ->label('Trạng thái phản hồi')
                    ->options(EnumPresenter::options(NotificationResponseStatus::class))
                    ->query(function (Builder $query, array $data): Builder {
                        $value = $data['value'] ?? null;

                        if (blank($value)) {
                            return $query;
                        }

                        return $query->where('notification_targets.response_status', (int) $value);
                    }),
            ])
            ->defaultSort('users.id', 'desc')
            ->headerActions([])
            ->recordActions([])
            ->toolbarActions([]);
    }
}
