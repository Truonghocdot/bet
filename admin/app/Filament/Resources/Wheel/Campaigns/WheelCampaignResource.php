<?php

namespace App\Filament\Resources\Wheel\Campaigns;

use App\Filament\Resources\BaseResource;
use App\Filament\Resources\Wheel\Campaigns\Pages\CreateWheelCampaign;
use App\Filament\Resources\Wheel\Campaigns\Pages\EditWheelCampaign;
use App\Filament\Resources\Wheel\Campaigns\Pages\ListWheelCampaigns;
use App\Models\User;
use App\Models\Wheel\WheelCampaign;
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

class WheelCampaignResource extends BaseResource
{
    protected static ?string $model = WheelCampaign::class;

    protected static UnitEnum|string|null $navigationGroup = 'Sự kiện vòng quay';

    protected static ?string $navigationLabel = 'Chiến dịch';

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedGift;

    public static function shouldRegisterNavigation(): bool
    {
        return true;
    }

    protected static function abilityPrefix(): string
    {
        return 'wheel.campaigns';
    }

    public static function form(Schema $schema): Schema
    {
        return $schema->components([
            TextInput::make('name')->label('Tên chiến dịch')->required()->maxLength(160)->columnSpanFull(),
            Select::make('status')->label('Trạng thái')->options(['draft' => 'Bản nháp', 'active' => 'Đang mở', 'closed' => 'Đã đóng'])->default('draft')->required(),
            Repeater::make('roundTemplates')
                ->label('Ba lượt mẫu')
                ->relationship()
                ->schema([
                    Select::make('round_no')->label('Lượt')->options([1 => 'Lượt 1', 2 => 'Lượt 2', 3 => 'Lượt 3'])->required()->disableOptionsWhenSelectedInSiblingRepeaterItems(),
                    TextInput::make('segment_key')->label('Mã ô')->required()->maxLength(64),
                    TextInput::make('result_label')->label('Kết quả hiển thị')->required()->maxLength(160),
                    TextInput::make('prize_amount')->label('Thưởng VND')->numeric()->minValue(0)->default(0)->required(),
                ])
                ->default([
                    ['round_no' => 1, 'segment_key' => 'try_again', 'result_label' => 'Chúc bạn may mắn', 'prize_amount' => 0],
                    ['round_no' => 2, 'segment_key' => 'reward_39m', 'result_label' => '39 triệu', 'prize_amount' => 39000000],
                    ['round_no' => 3, 'segment_key' => 'try_again', 'result_label' => 'Chúc bạn may mắn', 'prize_amount' => 0],
                ])
                ->helperText('Chiến dịch có đúng 3 lượt; lượt 2 luôn được khóa ở giải 39.000.000 VND.')
                ->minItems(3)->maxItems(3)->reorderable(false)->columns(4)->columnSpanFull(),
        ])->columns(2);
    }

    public static function table(Table $table): Table
    {
        return $table
            ->columns([
                TextColumn::make('id')->label('ID')->sortable(),
                TextColumn::make('name')->label('Chiến dịch')->searchable()->sortable(),
                TextColumn::make('status')->label('Trạng thái')->badge()->formatStateUsing(fn (string $state): string => match ($state) {
                    'active' => 'Đang mở', 'closed' => 'Đã đóng', default => 'Bản nháp'
                })->color(fn (string $state): string => match ($state) {
                    'active' => 'success', 'closed' => 'gray', default => 'warning'
                }),
                TextColumn::make('invitations_count')->label('Lời mời')->counts('invitations')->badge(),
                TextColumn::make('createdBy.name')->label('Người tạo'),
            ])
            ->recordActions([
                Action::make('invite_users')
                    ->label('Kích hoạt người chơi')->icon('heroicon-o-user-plus')->color('success')
                    ->visible(fn (WheelCampaign $record): bool => $record->status === 'active')
                    ->schema([
                        Toggle::make('bot_chat_enabled')
                            ->label('Bật bot chat cho người chơi')
                            ->default(true)
                            ->helperText('Tạo ngay 4 tin mở đầu, sau đó bot tiếp tục gửi khoảng 1 tin mỗi giây để người chơi thấy chat liên tục.')
                            ->columnSpanFull(),
                        Select::make('user_ids')
                            ->label('Người chơi')
                            ->multiple()
                            ->searchable()
                            ->required()
                            ->options(fn (Select $component): array => User::query()
                                ->whereIn('role', [2, 4])
                                ->whereIn('id', array_values(array_filter((array) $component->getState())))
                                ->get(['id', 'name', 'phone'])
                                ->mapWithKeys(fn (User $user): array => [$user->id => "#{$user->id} - {$user->name} - ".($user->phone ?: 'không có SĐT')])
                                ->all())
                            ->getSearchResultsUsing(fn (string $search): array => User::query()
                                ->whereIn('role', [2, 4])
                                ->where(fn (Builder $query) => $query->where('name', 'ilike', "%{$search}%")->orWhere('phone', 'ilike', "%{$search}%")->orWhereRaw('CAST(id AS TEXT) LIKE ?', ["%{$search}%"]))
                                ->limit(50)
                                ->get()
                                ->mapWithKeys(fn (User $user): array => [$user->id => "#{$user->id} - {$user->name} - ".($user->phone ?: 'không có SĐT')])
                                ->all())
                            ->getOptionLabelsUsing(fn (array $values): array => User::query()
                                ->whereIn('role', [2, 4])
                                ->whereIn('id', $values)
                                ->get(['id', 'name', 'phone'])
                                ->mapWithKeys(fn (User $user): array => [$user->id => "#{$user->id} - {$user->name} - ".($user->phone ?: 'không có SĐT')])
                                ->all()),
                    ])
                    ->action(function (WheelCampaign $record, array $data): void {
                        $count = app(WheelCampaignService::class)->inviteUsers($record, $data['user_ids'] ?? [], true, (bool) ($data['bot_chat_enabled'] ?? true));
                        Notification::make()->success()->title("Đã kích hoạt {$count} người chơi")->send();
                    }),
                EditAction::make(),
            ])
            ->headerActions([CreateAction::make()->label('Tạo chiến dịch')])
            ->defaultSort('id', 'desc');
    }

    public static function getPages(): array
    {
        return ['index' => ListWheelCampaigns::route('/'), 'create' => CreateWheelCampaign::route('/create'), 'edit' => EditWheelCampaign::route('/{record}/edit')];
    }
}
