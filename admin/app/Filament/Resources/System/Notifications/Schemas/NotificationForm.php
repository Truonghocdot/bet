<?php

namespace App\Filament\Resources\System\Notifications\Schemas;

use App\Enum\Notification\NotificationAudience;
use App\Enum\Notification\NotificationStatus;
use App\Models\User;
use App\Support\Filament\EnumPresenter;
use Filament\Forms\Components\DateTimePicker;
use Filament\Forms\Components\FileUpload;
use Filament\Forms\Components\RichEditor;
use Filament\Forms\Components\Select;
use Filament\Forms\Components\TextInput;
use Filament\Schemas\Components\Section;
use Filament\Schemas\Components\Utilities\Get;
use Filament\Schemas\Schema;

class NotificationForm
{
    public static function configure(Schema $schema): Schema
    {
        return $schema->components([
            Section::make('Nội dung thông báo')
                ->schema([
                    TextInput::make('title')
                        ->label('Tiêu đề')
                        ->required()
                        ->maxLength(200),
                    FileUpload::make('image_path')
                        ->label('Ảnh thông báo')
                        ->disk('public')
                        ->directory('notifications')
                        ->image()
                        ->imageEditor()
                        ->helperText('Nếu có ảnh và gửi cho người dùng chỉ định, app sẽ hiển thị nút Xác nhận / Hủy cho khách.')
                        ->columnSpanFull(),
                    RichEditor::make('body')
                        ->label('Nội dung')
                        ->required(fn (Get $get): bool => blank($get('image_path')))
                        ->toolbarButtons([
                            'blockquote',
                            'bold',
                            'bulletList',
                            'italic',
                            'link',
                            'orderedList',
                            'redo',
                            'strike',
                            'underline',
                            'undo',
                        ])
                        ->helperText('Có thể chèn link; app sẽ hiển thị nội dung HTML và người dùng có thể bấm được liên kết. Nếu đã có ảnh, nội dung này là caption tùy chọn.')
                        ->columnSpanFull(),
                ])
                ->columns(2),

            Section::make('Thiết lập phát hành')
                ->schema([
                    Select::make('status')
                        ->label('Trạng thái')
                        ->options(EnumPresenter::options(NotificationStatus::class))
                        ->default(NotificationStatus::DRAFT->value)
                        ->required(),
                    Select::make('audience')
                        ->label('Đối tượng nhận')
                        ->options(EnumPresenter::options(NotificationAudience::class))
                        ->default(NotificationAudience::ALL->value)
                        ->required()
                        ->live(),
                    DateTimePicker::make('publish_at')
                        ->label('Thời gian phát hành')
                        ->seconds(false)
                        ->timezone(config('app.timezone', 'Asia/Ho_Chi_Minh')),
                    DateTimePicker::make('expires_at')
                        ->label('Hết hạn lúc')
                        ->seconds(false)
                        ->timezone(config('app.timezone', 'Asia/Ho_Chi_Minh')),
                    Select::make('targetUsers')
                        ->label('Người dùng chỉ định')
                        ->multiple()
                        ->searchable()
                        ->preload()
                        ->options(fn (): array => User::query()
                            ->orderBy('phone')
                            ->limit(50)
                            ->get(['id', 'name', 'phone'])
                            ->mapWithKeys(static fn (User $user): array => [
                                $user->getKey() => trim(implode(' - ', array_filter([
                                    $user->phone,
                                    $user->name,
                                ]))),
                            ])
                            ->all())
                        ->getSearchResultsUsing(fn (string $search): array => User::query()
                            ->where(function ($query) use ($search): void {
                                $query
                                    ->where('phone', 'like', '%'.$search.'%')
                                    ->orWhere('name', 'like', '%'.$search.'%');
                            })
                            ->orderBy('phone')
                            ->limit(50)
                            ->get(['id', 'name', 'phone'])
                            ->mapWithKeys(static fn (User $user): array => [
                                $user->getKey() => trim(implode(' - ', array_filter([
                                    $user->phone,
                                    $user->name,
                                ]))),
                            ])
                            ->all())
                        ->getOptionLabelsUsing(fn (array $values): array => User::query()
                            ->whereIn('id', $values)
                            ->get(['id', 'name', 'phone'])
                            ->mapWithKeys(static fn (User $user): array => [
                                $user->getKey() => trim(implode(' - ', array_filter([
                                    $user->phone,
                                    $user->name,
                                ]))),
                            ])
                            ->all())
                        ->helperText('Chỉ áp dụng khi đối tượng nhận là "Người dùng chỉ định".')
                        ->visible(fn (Get $get): bool => (int) $get('audience') === NotificationAudience::USERS->value)
                        ->required(fn (Get $get): bool => (int) $get('audience') === NotificationAudience::USERS->value),
                ])
                ->columns(2),
        ]);
    }
}
