<?php

namespace App\Filament\Resources\Chat\ChatMessages\Tables;

use App\Models\Chat\ChatBan;
use App\Models\Chat\ChatMessage;
use App\Models\Chat\ChatModerationAction;
use App\Services\Chat\ChatRedisPublisher;
use Filament\Actions\Action;
use Filament\Forms\Components\Textarea;
use Filament\Notifications\Notification;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Table;
use Illuminate\Support\Facades\Gate;

class ChatMessagesTable
{
    public static function configure(Table $table): Table
    {
        return $table->columns([
            TextColumn::make('id')->label('ID')->sortable(),
            TextColumn::make('display_name')->label('Tên chat')->searchable(),
            TextColumn::make('actor_type')->label('Nguồn')->badge(),
            TextColumn::make('body')->label('Nội dung')->limit(80)->wrap(),
            TextColumn::make('status')->label('Trạng thái')->formatStateUsing(fn ($state): string => match ((int) $state) {
                1 => 'Hiển thị', 2 => 'Ẩn', 3 => 'Đã xóa', default => 'Không rõ'
            }),
            TextColumn::make('created_at')->label('Thời gian')->dateTime('d/m/Y H:i')->sortable(),
        ])->defaultSort('id', 'desc')->poll(5000)->recordActions([
            Action::make('hide')->label('Ẩn')->icon('heroicon-o-eye-slash')->color('warning')
                ->visible(fn (ChatMessage $record): bool => Gate::allows('chat.messages.update') && $record->status === ChatMessage::STATUS_VISIBLE)
                ->schema([Textarea::make('reason')->label('Lý do')->maxLength(255)])
                ->action(fn (ChatMessage $record, array $data) => self::moderate($record, 'hidden', $data['reason'] ?? null)),
            Action::make('delete')->label('Xóa')->icon('heroicon-o-trash')->color('danger')->requiresConfirmation()
                ->visible(fn (ChatMessage $record): bool => Gate::allows('chat.messages.delete') && $record->status !== ChatMessage::STATUS_DELETED)
                ->schema([Textarea::make('reason')->label('Lý do')->maxLength(255)])
                ->action(fn (ChatMessage $record, array $data) => self::moderate($record, 'deleted', $data['reason'] ?? null)),
            Action::make('ban')->label('Khóa chat user')->icon('heroicon-o-no-symbol')->color('danger')->requiresConfirmation()
                ->visible(fn (ChatMessage $record): bool => Gate::allows('chat.moderation.update') && $record->user_id)
                ->schema([Textarea::make('reason')->label('Lý do')->maxLength(255)])
                ->action(function (ChatMessage $record, array $data): void {
                    ChatBan::query()->create(['room_id' => $record->room_id, 'user_id' => $record->user_id, 'created_by' => auth()->id(), 'reason' => $data['reason'] ?? null]);
                    ChatModerationAction::query()->create(['actor_user_id' => auth()->id(), 'target_user_id' => $record->user_id, 'message_id' => $record->id, 'action' => 'ban', 'reason' => $data['reason'] ?? null]);
                    Notification::make()->title('Đã khóa quyền chat')->success()->send();
                }),
        ]);
    }

    private static function moderate(ChatMessage $record, string $action, ?string $reason): void
    {
        $status = $action === 'hidden' ? ChatMessage::STATUS_HIDDEN : ChatMessage::STATUS_DELETED;
        $record->forceFill(['status' => $status, 'moderated_by' => auth()->id(), 'moderated_at' => now()])->save();
        ChatModerationAction::query()->create(['actor_user_id' => auth()->id(), 'target_user_id' => $record->user_id, 'message_id' => $record->id, 'action' => $action, 'reason' => $reason]);
        try {
            $record->loadMissing('room');
            if ($record->room?->wheel_session_id) {
                app(ChatRedisPublisher::class)->publishWheelSession((int) $record->room->wheel_session_id, 'chat.message.'.($action === 'hidden' ? 'hidden' : 'deleted'), $record);
            } elseif ($record->room?->wheel_invitation_id) {
                app(ChatRedisPublisher::class)->publishWheelInvitation((int) $record->room->wheel_invitation_id, 'chat.message.'.($action === 'hidden' ? 'hidden' : 'deleted'), $record);
            } else {
                app(ChatRedisPublisher::class)->publish('global', 'chat.message.'.($action === 'hidden' ? 'hidden' : 'deleted'), $record);
            }
        } catch (\Throwable) { /* REST reflects the durable status. */
        }
        Notification::make()->title($action === 'hidden' ? 'Đã ẩn tin nhắn' : 'Đã xóa tin nhắn')->success()->send();
    }
}
