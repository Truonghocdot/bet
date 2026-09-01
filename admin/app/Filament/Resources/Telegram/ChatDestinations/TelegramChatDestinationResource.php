<?php

namespace App\Filament\Resources\Telegram\ChatDestinations;

use App\Filament\Resources\BaseResource;
use App\Filament\Resources\Telegram\ChatDestinations\Pages\ListTelegramChatDestinations;
use App\Models\Telegram\TelegramChatDestination;
use App\Services\Telegram\TelegramChatDestinationService;
use BackedEnum;
use Filament\Actions\Action;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Columns\IconColumn;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Table;
use Illuminate\Database\Eloquent\Builder;
use UnitEnum;

class TelegramChatDestinationResource extends BaseResource
{
    protected static ?string $model = TelegramChatDestination::class;

    protected static bool $canCreateRecords = false;
    protected static bool $canDeleteRecords = false;
    protected static bool $canDeleteAnyRecords = false;

    protected static UnitEnum|string|null $navigationGroup = 'Thông báo';
    protected static ?string $navigationLabel = 'Telegram Groups';
    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedPaperAirplane;

    public static function shouldRegisterNavigation(): bool
    {
        return true;
    }

    protected static function abilityPrefix(): string
    {
        return 'telegram.chat-destinations';
    }

    public static function getEloquentQuery(): Builder
    {
        return parent::getEloquentQuery()->where('site_code', static::siteCode());
    }

    public static function table(Table $table): Table
    {
        return $table
            ->columns([
                TextColumn::make('id')->label('ID')->sortable(),
                TextColumn::make('title')->label('Tên group')->searchable()->default('—'),
                TextColumn::make('username')->label('Username')->formatStateUsing(fn (?string $state): string => $state ? '@'.$state : '—'),
                TextColumn::make('telegram_chat_id')->label('Chat ID')->copyable()->fontFamily('mono'),
                TextColumn::make('bot_status')->label('Trạng thái bot')->badge(),
                IconColumn::make('is_active')->label('Nhận thông báo')->boolean(),
                TextColumn::make('last_error')->label('Lỗi gần nhất')->limit(60)->toggleable(),
                TextColumn::make('last_seen_at')->label('Nhìn thấy lần cuối')->dateTime('d/m/Y H:i:s', timezone: config('app.timezone'))->sortable(),
                TextColumn::make('activated_at')->label('Active lúc')->dateTime('d/m/Y H:i:s', timezone: config('app.timezone'))->toggleable(),
            ])
            ->defaultSort('id', 'desc')
            ->poll(5000)
            ->recordActions([
                Action::make('activate')
                    ->label('Active')
                    ->icon('heroicon-o-check-circle')
                    ->color('success')
                    ->requiresConfirmation()
                    ->visible(fn (TelegramChatDestination $record): bool => self::canEdit($record) && ! $record->is_active)
                    ->action(function (TelegramChatDestination $record, TelegramChatDestinationService $service): void {
                        $service->activate($record);
                    }),
                Action::make('deactivate')
                    ->label('Deactive')
                    ->icon('heroicon-o-x-circle')
                    ->color('warning')
                    ->requiresConfirmation()
                    ->visible(fn (TelegramChatDestination $record): bool => self::canEdit($record) && $record->is_active)
                    ->action(function (TelegramChatDestination $record, TelegramChatDestinationService $service): void {
                        $service->deactivate($record);
                    }),
                Action::make('send_test')
                    ->label('Gửi tin thử')
                    ->icon('heroicon-o-paper-airplane')
                    ->color('info')
                    ->visible(fn (TelegramChatDestination $record): bool => self::canEdit($record) && $record->is_active)
                    ->action(function (TelegramChatDestination $record, TelegramChatDestinationService $service): void {
                        $service->sendTest($record);
                    }),
            ]);
    }

    public static function getPages(): array
    {
        return ['index' => ListTelegramChatDestinations::route('/')];
    }

    private static function siteCode(): string
    {
        return trim((string) config('services.telegram.site_code', env('WHEEL_SITE_CODE', 'fh88u')));
    }
}
