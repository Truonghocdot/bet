<?php

namespace App\Filament\Resources\Wheel\Invitations;

use App\Enum\User\RoleUser;
use App\Filament\Resources\BaseResource;
use App\Filament\Resources\Wheel\Invitations\Pages\CreateWheelInvitation;
use App\Filament\Resources\Wheel\Invitations\Pages\EditWheelInvitation;
use App\Filament\Resources\Wheel\Invitations\Pages\ListWheelInvitations;
use App\Models\User;
use App\Models\Wheel\WheelInvitation;
use App\Services\Wheel\WheelCampaignService;
use BackedEnum;
use Filament\Actions\Action;
use Filament\Actions\CreateAction;
use Filament\Actions\EditAction;
use Filament\Forms\Components\Repeater;
use Filament\Forms\Components\Select;
use Filament\Forms\Components\TextInput;
use Filament\Forms\Components\Toggle;
use Filament\Notifications\Notification;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Filament\Tables\Columns\TextColumn;
use Filament\Tables\Table;
use Illuminate\Database\Eloquent\Builder;
use UnitEnum;

class WheelInvitationResource extends BaseResource
{
    protected static ?string $model = WheelInvitation::class;

    protected static UnitEnum|string|null $navigationGroup = 'Sự kiện vòng quay';

    protected static ?string $navigationLabel = 'Người được mời';

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedTicket;

    public static function shouldRegisterNavigation(): bool
    {
        return true;
    }

    protected static function abilityPrefix(): string
    {
        return 'wheel.invitations';
    }

    public static function canCreate(): bool
    {
        return in_array(auth()->user()?->role, [RoleUser::SUPER_ADMIN, RoleUser::ADMIN], true);
    }

    public static function canEdit($record): bool
    {
        return in_array(auth()->user()?->role, [RoleUser::SUPER_ADMIN, RoleUser::ADMIN], true);
    }

    public static function form(Schema $schema): Schema
    {
        return $schema->components([
            Select::make('campaign_id')
                ->label('Chiến dịch')
                ->relationship('campaign', 'name')
                ->placeholder('Chọn chiến dịch')
                ->searchable()
                ->searchPrompt('Tìm chiến dịch...')
                ->noSearchResultsMessage('Không tìm thấy chiến dịch')
                ->preload()
                ->optionsLimit(50)
                ->disabled(fn (?WheelInvitation $record): bool => $record?->status !== null && $record->status !== 'draft'),
            Select::make('user_id')->label('Người chơi')->required()->searchable()
                ->getSearchResultsUsing(fn (string $search): array => User::query()->whereIn('role', [2, 4])->where(fn (Builder $query) => $query->where('name', 'ilike', "%{$search}%")->orWhere('phone', 'ilike', "%{$search}%")->orWhereRaw('CAST(id AS TEXT) LIKE ?', ["%{$search}%"]))->limit(50)->get()->mapWithKeys(fn (User $user): array => [$user->id => "#{$user->id} - {$user->name} - ".($user->phone ?: 'không có SĐT')])->all())
                ->getOptionLabelUsing(fn ($value): string => optional(User::find($value), fn (User $user): string => "#{$user->id} - {$user->name} - ".($user->phone ?: 'không có SĐT')) ?? (string) $value)
                ->disabled(fn (?WheelInvitation $record): bool => $record?->status !== null && $record->status !== 'draft'),
            Toggle::make('bot_chat_enabled')
                ->label('Bật bot chat cho người chơi')
                ->default(true)
                ->helperText('Tạo ngay 4 tin mở đầu, sau đó bot tiếp tục gửi khoảng 1 tin mỗi giây.')
                ->disabled(fn (?WheelInvitation $record): bool => $record?->status !== null && $record->status !== 'draft'),
            TextInput::make('status')->label('Trạng thái')->disabled()->dehydrated(false),
            Repeater::make('rounds')->label('Kết quả riêng của người chơi')->relationship()->schema([
                Select::make('round_no')->label('Lượt')->options([1 => 'Lượt 1', 2 => 'Lượt 2', 3 => 'Lượt 3'])->required()->disableOptionsWhenSelectedInSiblingRepeaterItems(),
                TextInput::make('segment_key')->label('Mã ô')->required()->maxLength(64),
                TextInput::make('result_label')->label('Kết quả')->required()->maxLength(160),
                TextInput::make('prize_amount')->label('Thưởng VND')->numeric()->minValue(0)->required(),
            ])->helperText('Chỉ có 3 lượt; lượt 2 luôn được lưu là giải 39.000.000 VND.')->minItems(3)->maxItems(3)->reorderable(false)->columns(4)->columnSpanFull()->hiddenOn('create')->disabled(fn (?WheelInvitation $record): bool => $record?->status !== 'draft'),
        ])->columns(2);
    }

    public static function table(Table $table): Table
    {
        return $table->columns([
            TextColumn::make('id')->label('ID')->sortable(),
            TextColumn::make('campaign.name')->label('Chiến dịch')->searchable(),
            TextColumn::make('user.id')->label('User ID')->sortable(),
            TextColumn::make('user.name')->label('Người chơi')->searchable(),
            TextColumn::make('user.phone')->label('Số điện thoại')->searchable(),
            TextColumn::make('status')->label('Trạng thái')->badge(),
            TextColumn::make('bot_chat_enabled')
                ->label('Bot chat')
                ->badge()
                ->formatStateUsing(fn (bool $state): string => $state ? 'Đang bật' : 'Đang tắt')
                ->color(fn (bool $state): string => $state ? 'success' : 'gray'),
            TextColumn::make('session.current_round')->label('Lượt hiện tại')->default('—'),
            TextColumn::make('activated_at')->label('Kích hoạt lúc')->dateTime('d/m/Y H:i', timezone: config('app.timezone')),
        ])->recordActions([
            Action::make('activate')->label('Kích hoạt')->icon('heroicon-o-bolt')->color('success')->requiresConfirmation()->visible(fn (WheelInvitation $record): bool => self::canEdit($record) && $record->status === 'draft')->action(function (WheelInvitation $record): void {
                app(WheelCampaignService::class)->activate($record);
                Notification::make()->success()->title('Đã kích hoạt lời mời')->send();
            }),
            Action::make('enable_bot_chat')->label('Bật bot chat')->icon('heroicon-o-chat-bubble-left-right')->color('success')->requiresConfirmation()
                ->visible(fn (WheelInvitation $record): bool => self::canEdit($record) && ! $record->bot_chat_enabled && in_array($record->status, ['pending', 'started'], true))
                ->action(function (WheelInvitation $record): void {
                    app(WheelCampaignService::class)->setBotChatEnabled($record, true);
                    Notification::make()->success()->title('Đã bật bot và khởi động phòng chat')->send();
                }),
            Action::make('disable_bot_chat')->label('Tắt bot chat')->icon('heroicon-o-chat-bubble-left-ellipsis')->color('warning')->requiresConfirmation()
                ->visible(fn (WheelInvitation $record): bool => self::canEdit($record) && $record->bot_chat_enabled && in_array($record->status, ['pending', 'started'], true))
                ->action(function (WheelInvitation $record): void {
                    app(WheelCampaignService::class)->setBotChatEnabled($record, false);
                    Notification::make()->success()->title('Đã dừng bot chat')->send();
                }),
            Action::make('revoke')->label('Thu hồi')->icon('heroicon-o-x-circle')->color('danger')->requiresConfirmation()->visible(fn (WheelInvitation $record): bool => self::canEdit($record) && in_array($record->status, ['draft', 'pending'], true))->action(function (WheelInvitation $record): void {
                app(WheelCampaignService::class)->revoke($record);
                Notification::make()->success()->title('Đã thu hồi lời mời')->send();
            }),
            EditAction::make(),
        ])->headerActions([CreateAction::make()->label('Tạo lời mời tùy chỉnh')])->defaultSort('id', 'desc')->poll(3000);
    }

    public static function getPages(): array
    {
        return ['index' => ListWheelInvitations::route('/'), 'create' => CreateWheelInvitation::route('/create'), 'edit' => EditWheelInvitation::route('/{record}/edit')];
    }
}
