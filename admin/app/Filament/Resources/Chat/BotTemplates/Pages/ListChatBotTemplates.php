<?php

namespace App\Filament\Resources\Chat\BotTemplates\Pages;

use App\Filament\Resources\Chat\BotTemplates\ChatBotTemplateResource;
use App\Services\Chat\ChatBotBulkImportService;
use Filament\Actions\Action;
use Filament\Actions\CreateAction;
use Filament\Forms\Components\Textarea;
use Filament\Forms\Components\TextInput;
use Filament\Forms\Components\Toggle;
use Filament\Notifications\Notification;
use Filament\Resources\Pages\ListRecords;
use Filament\Support\Enums\Width;

class ListChatBotTemplates extends ListRecords
{
    protected static string $resource = ChatBotTemplateResource::class;

    protected function getHeaderActions(): array
    {
        return [
            Action::make('paste_bulk_templates')
                ->label('Dán câu mẫu hàng loạt')
                ->icon('heroicon-o-clipboard-document-list')
                ->color('success')
                ->visible(fn (): bool => ChatBotTemplateResource::canCreate())
                ->modalHeading('Dán danh sách bot và câu mẫu')
                ->modalDescription('Mỗi dòng có thể là “người chơi# 258425[TAB]Nội dung” hoặc chỉ có nội dung. Hệ thống tự bỏ qua câu trùng.')
                ->modalWidth(Width::FiveExtraLarge)
                ->modalSubmitActionLabel('Import danh sách')
                ->schema([
                    Textarea::make('bulk_text')
                        ->label('Danh sách cần import')
                        ->required()
                        ->rows(20)
                        ->maxLength(200000)
                        ->helperText('Có thể dán trực tiếp toàn bộ nội dung file bot chat.txt. Giữ nguyên mỗi bot và câu nói trên một dòng.'),
                    TextInput::make('category')->label('Nhóm nội dung')->default('event')->required()->maxLength(60),
                    TextInput::make('language')->label('Ngôn ngữ')->default('vi')->required()->maxLength(12),
                    Toggle::make('active')->label('Bật các câu vừa import')->default(true),
                    Toggle::make('only_game_ids')
                        ->label('Chỉ dùng profile dạng ID game')
                        ->helperText('Tắt các profile tên cũ sau khi import để chat chỉ hiển thị ID game.')
                        ->default(true),
                ])
                ->action(function (array $data): void {
                    $result = app(ChatBotBulkImportService::class)->import(
                        (string) $data['bulk_text'],
                        (string) ($data['category'] ?? 'event'),
                        (string) ($data['language'] ?? 'vi'),
                        (bool) ($data['active'] ?? true),
                        (bool) ($data['only_game_ids'] ?? true),
                    );

                    Notification::make()
                        ->title('Đã import danh sách câu mẫu')
                        ->body(sprintf(
                            '%d câu mới, %d profile mới, %d profile cập nhật, %d câu trùng, %d dòng lỗi, %d profile tên cũ đã tắt.',
                            $result['templates_created'],
                            $result['profiles_created'],
                            $result['profiles_updated'],
                            $result['duplicates'],
                            $result['invalid'],
                            $result['profiles_deactivated'],
                        ))
                        ->success()
                        ->send();
                }),
            CreateAction::make()->label('Thêm một câu'),
        ];
    }
}
