<?php

namespace App\Filament\Resources\Chat\ModerationActions;

use App\Filament\Resources\BaseResource;
use App\Filament\Resources\Chat\ModerationActions\Pages\ListChatModerationActions;
use App\Models\Chat\ChatModerationAction;
use BackedEnum;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Table;
use UnitEnum;

class ChatModerationActionResource extends BaseResource
{
    protected static ?string $model = ChatModerationAction::class;

    protected static bool $canCreateRecords = false;

    protected static bool $canUpdateRecords = false;

    protected static bool $canDeleteRecords = false;

    protected static bool $canDeleteAnyRecords = false;

    protected static UnitEnum|string|null $navigationGroup = 'Chat';

    protected static ?string $navigationLabel = 'Audit moderation';

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedClipboardDocumentList;

    public static function shouldRegisterNavigation(): bool
    {
        return true;
    }

    protected static function abilityPrefix(): string
    {
        return 'chat.moderation-actions';
    }

    public static function table(Table $table): Table
    {
        return $table->columns([
            TextColumn::make('id')->label('ID')->sortable(),
            TextColumn::make('action')->label('Hành động')->badge(),
            TextColumn::make('actor.name')->label('Nhân viên'),
            TextColumn::make('target_user_id')->label('ID user'),
            TextColumn::make('message_id')->label('ID tin'),
            TextColumn::make('reason')->label('Lý do')->limit(80),
            TextColumn::make('created_at')->label('Thời gian')->dateTime('d/m/Y H:i')->sortable(),
        ])->defaultSort('id', 'desc');
    }

    public static function getPages(): array
    {
        return ['index' => ListChatModerationActions::route('/')];
    }
}
