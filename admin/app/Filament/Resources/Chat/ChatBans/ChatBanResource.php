<?php

namespace App\Filament\Resources\Chat\ChatBans;

use App\Filament\Resources\BaseResource;
use App\Filament\Resources\Chat\ChatBans\Pages\ListChatBans;
use App\Models\Chat\ChatBan;
use App\Models\Chat\ChatModerationAction;
use BackedEnum;
use Filament\Actions\Action;
use Filament\Notifications\Notification;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Table;
use UnitEnum;

class ChatBanResource extends BaseResource
{
    protected static ?string $model = ChatBan::class;

    protected static bool $canCreateRecords = false;

    protected static bool $canUpdateRecords = false;

    protected static bool $canDeleteRecords = false;

    protected static bool $canDeleteAnyRecords = false;

    protected static UnitEnum|string|null $navigationGroup = 'Chat';

    protected static ?string $navigationLabel = 'Khóa chat';

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedNoSymbol;

    public static function shouldRegisterNavigation(): bool
    {
        return true;
    }

    protected static function abilityPrefix(): string
    {
        return 'chat.bans';
    }

    public static function table(Table $table): Table
    {
        return $table->columns([
            TextColumn::make('id')->label('ID')->sortable(),
            TextColumn::make('user.id')->label('ID user')->sortable(),
            TextColumn::make('user.name')->label('Người dùng')->searchable(),
            TextColumn::make('reason')->label('Lý do')->limit(60),
            TextColumn::make('created_at')->label('Khóa lúc')->dateTime('d/m/Y H:i')->sortable(),
            TextColumn::make('revoked_at')->label('Mở lúc')->dateTime('d/m/Y H:i'),
        ])->defaultSort('id', 'desc')->recordActions([
            Action::make('unban')->label('Mở khóa')->icon('heroicon-o-lock-open')->color('success')
                ->visible(fn (ChatBan $record): bool => $record->revoked_at === null)
                ->requiresConfirmation()
                ->action(function (ChatBan $record): void {
                    $record->forceFill(['revoked_at' => now(), 'revoked_by' => auth()->id()])->save();
                    ChatModerationAction::query()->create(['actor_user_id' => auth()->id(), 'target_user_id' => $record->user_id, 'action' => 'unban']);
                    Notification::make()->title('Đã mở khóa chat')->success()->send();
                }),
        ]);
    }

    public static function getPages(): array
    {
        return ['index' => ListChatBans::route('/')];
    }
}
