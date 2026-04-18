<?php

namespace App\Filament\Resources\Bet\GamePeriods\Tables;

use App\Enum\Bet\DrawSource;
use App\Enum\Bet\GameType;
use App\Enum\Bet\PeriodStatus;
use App\Models\Bet\GameRoom;
use App\Models\User;
use App\Support\Filament\EnumPresenter;
use Filament\Actions\Action;
use Filament\Actions\ActionGroup;
use Filament\Actions\EditAction;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Table;
use Illuminate\Support\Facades\Gate;
use Illuminate\Support\Facades\Redis;
use Illuminate\Support\Str;
use Throwable;

class GamePeriodsTable
{
    public static function configure(Table $table): Table
    {
        return $table
            ->columns([
                TextColumn::make('id')->label('ID')->sortable(),
                TextColumn::make('game_type')
                    ->label('Trò chơi')
                    ->badge()
                    ->formatStateUsing(fn ($state): string => EnumPresenter::label(GameType::class, $state))
                    ->color(fn ($state): string => EnumPresenter::color(GameType::class, $state)),
                TextColumn::make('period_no')->label('Kỳ số')->searchable()->sortable(),
                TextColumn::make('period_index')
                    ->label('Mã kỳ')
                    ->sortable()
                    ->searchable()
                    ->copyable()
                    ->fontFamily('mono'),
                TextColumn::make('room_code')->label('Phòng')->toggleable(),
                TextColumn::make('status')
                    ->label('Trạng thái')
                    ->badge()
                    ->formatStateUsing(fn ($state): string => EnumPresenter::label(PeriodStatus::class, $state))
                    ->color(fn ($state): string => EnumPresenter::color(PeriodStatus::class, $state)),
                TextColumn::make('draw_source')
                    ->label('Nguồn')
                    ->badge()
                    ->formatStateUsing(fn ($state): string => EnumPresenter::label(DrawSource::class, $state))
                    ->color(fn ($state): string => EnumPresenter::color(DrawSource::class, $state)),
                TextColumn::make('open_at')->label('Mở lúc')->dateTime()->sortable()->toggleable(),
                TextColumn::make('close_at')->label('Đóng lúc')->dateTime()->sortable()->toggleable(),
                TextColumn::make('draw_at')->label('Quay lúc')->dateTime()->sortable()->toggleable(),
                TextColumn::make('settled_at')->label('Chốt lúc')->dateTime()->toggleable(),
                TextColumn::make('created_at')->label('Tạo lúc')->dateTime()->sortable(),
            ])
            ->defaultSort('id', 'desc')
            ->poll(2000)
            ->headerActions(static::controlRoomActions());
    }

    /**
     * @return array<int, Action|ActionGroup>
     */
    protected static function controlRoomActions(): array
    {
        if (! Gate::allows('control_panel_access')) {
            return [];
        }

        $presence = static::controlRoomPresence();
        $userLabels = static::controlRoomUserLabels($presence);
        $groups = [];
        $roomsByGame = GameRoom::query()
            ->orderBy('sort_order')
            ->get()
            ->groupBy(fn (GameRoom $room): int => (int) $room->game_type->value);

        foreach ($roomsByGame as $gameType => $rooms) {
            $actions = [];

            foreach ($rooms as $room) {
                $roomCode = (string) $room->code;
                $users = $presence[$roomCode] ?? [];
                $count = count($users);
                $isBusy = $count > 0;
                $slotLabel = $isBusy ? sprintf('%d admin đang vào', $count) : 'Slot trống';
                $tooltip = $isBusy
                    ? sprintf(
                        '%s đang có: %s',
                        $room->controlLabel(),
                        implode(', ', array_map(
                            static fn (int $userID): string => $userLabels[$userID] ?? 'Admin #'.$userID,
                            $users,
                        )),
                    )
                    : sprintf('%s hiện chưa có admin nào trong control room', $room->controlLabel());

                $actions[] = Action::make('open_control_room_'.$roomCode)
                    ->label($room->controlLabel().' • '.$slotLabel)
                    ->icon($isBusy ? 'heroicon-m-signal' : 'heroicon-m-bolt')
                    ->color($isBusy ? 'danger' : 'success')
                    ->tooltip($tooltip)
                    ->url(route('auth.sso.redirect', ['room_code' => $roomCode]))
                    ->openUrlInNewTab();
            }

            $groups[] = ActionGroup::make($actions)
                ->label(static::controlGroupLabel((int) $gameType))
                ->icon(static::controlGroupIcon((int) $gameType))
                ->color(static::controlGroupColor((int) $gameType))
                ->button();
        }

        return $groups;
    }

    /**
     * @return array<string, array<int, int>>
     */
    protected static function controlRoomPresence(): array
    {
        $presence = [];

        try {
            $redis = Redis::connection('shared');
            $keys = $redis->keys(GameRoom::CONTROL_LOCK_REDIS_PREFIX.'*');

            foreach ($keys as $key) {
                $roomCode = trim((string) Str::after((string) $key, GameRoom::CONTROL_LOCK_REDIS_PREFIX));
                if ($roomCode === '') {
                    continue;
                }

                $payload = json_decode((string) $redis->get($key), true);
                $userID = (int) data_get($payload, 'user_id', 0);
                if ($roomCode === '' || $userID <= 0) {
                    continue;
                }

                $presence[$roomCode] ??= [];
                if (! in_array($userID, $presence[$roomCode], true)) {
                    $presence[$roomCode][] = $userID;
                }
            }
        } catch (Throwable) {
            return [];
        }

        return $presence;
    }

    /**
     * @param  array<string, array<int, int>>  $presence
     * @return array<int, string>
     */
    protected static function controlRoomUserLabels(array $presence): array
    {
        $userIDs = collect($presence)
            ->flatten()
            ->map(fn ($userID): int => (int) $userID)
            ->filter(fn (int $userID): bool => $userID > 0)
            ->unique()
            ->values()
            ->all();

        if ($userIDs === []) {
            return [];
        }

        return User::query()
            ->whereIn('id', $userIDs)
            ->get(['id', 'name'])
            ->mapWithKeys(function (User $user): array {
                $name = trim((string) $user->name);

                return [
                    $user->id => $name !== ''
                        ? sprintf('%s (#%d)', $name, $user->id)
                        : 'Admin #'.$user->id,
                ];
            })
            ->all();
    }

    protected static function controlGroupLabel(int $gameType): string
    {
        return match ($gameType) {
            GameType::WINGO->value => 'Control Wingo',
            GameType::K3->value => 'Control K3',
            GameType::LOTTERY->value => 'Control 5D',
            default => 'Control Room',
        };
    }

    protected static function controlGroupIcon(int $gameType): string
    {
        return match ($gameType) {
            GameType::WINGO->value => 'heroicon-m-fire',
            GameType::K3->value => 'heroicon-m-squares-2x2',
            GameType::LOTTERY->value => 'heroicon-m-sparkles',
            default => 'heroicon-m-computer-desktop',
        };
    }

    protected static function controlGroupColor(int $gameType): string
    {
        return match ($gameType) {
            GameType::WINGO->value => 'danger',
            GameType::K3->value => 'warning',
            GameType::LOTTERY->value => 'info',
            default => 'gray',
        };
    }
}
