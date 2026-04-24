<?php

namespace App\Filament\Resources\Users\Schemas;

use App\Enum\User\RoleUser;
use App\Enum\User\UserStatus;
use App\Enum\Wallet\UnitTransaction;
use App\Support\Filament\EnumPresenter;
use Filament\Forms\Components\DateTimePicker;
use Filament\Forms\Components\Hidden;
use Filament\Forms\Components\Select;
use Filament\Forms\Components\TextInput;
use Filament\Forms\Components\Toggle;
use Filament\Schemas\Components\Section;
use Filament\Schemas\Components\Utilities\Get;
use Filament\Schemas\Schema;

class UserForm
{
    public static function roleOptionsForCurrentActor(): array
    {
        $actorRole = auth()->user()?->role;
        $allowedValues = RoleUser::manageableValuesBy($actorRole instanceof RoleUser ? $actorRole : null);

        return collect(EnumPresenter::options(RoleUser::class))
            ->only($allowedValues)
            ->all();
    }

    public static function configure(Schema $schema, ?RoleUser $fixedRole = null): Schema
    {
        $roleField = $fixedRole
            ? Hidden::make('role')->default($fixedRole->value)
            : Select::make('role')
                ->label('Vai trò')
                ->options(static::roleOptionsForCurrentActor())
                ->required();

        return $schema
            ->components([
                Section::make('Thông tin cơ bản')
                    ->schema([
                        TextInput::make('name')
                            ->label('Họ và tên')
                            ->required()
                            ->maxLength(100),
                        TextInput::make('phone')
                            ->label('Số điện thoại')
                            ->tel()
                            ->maxLength(20)
                            ->unique(ignoreRecord: true)
                            ->helperText('Có thể để trống nếu chưa có số điện thoại.'),
                    ])
                    ->columns(2),
                Section::make('Mật khẩu đăng nhập')
                    ->description('Tạo mới có thể nhập mật khẩu ngay. Khi chỉnh sửa, chỉ cần bật đổi mật khẩu rồi nhập mật khẩu mới.')
                    ->schema([
                        Toggle::make('change_password')
                            ->label('Bật thay đổi mật khẩu')
                            ->default(false)
                            ->dehydrated(false)
                            ->live()
                            ->visible(fn (string $operation): bool => $operation === 'edit'),
                        TextInput::make('password')
                            ->label(fn (string $operation): string => $operation === 'create' ? 'Mật khẩu đăng nhập' : 'Mật khẩu mới')
                            ->password()
                            ->revealable()
                            ->autocomplete('new-password')
                            ->required(fn (Get $get, string $operation): bool => $operation === 'create' || (bool) $get('change_password'))
                            ->visible(fn (Get $get, string $operation): bool => $operation === 'create' || (bool) $get('change_password'))
                            ->dehydrated(fn (?string $state, Get $get, string $operation): bool => filled($state) && ($operation === 'create' || (bool) $get('change_password')))
                            ->maxLength(255)
                            ->helperText(fn (string $operation): string => $operation === 'create'
                                ? 'Nhập mật khẩu ngay khi tạo tài khoản đại lý/người dùng.'
                                : 'Bật đổi mật khẩu để nhập mật khẩu mới cho tài khoản này.'),
                        TextInput::make('password_confirmation')
                            ->label('Nhập lại mật khẩu')
                            ->password()
                            ->revealable()
                            ->autocomplete('new-password')
                            ->same('password')
                            ->required(fn (Get $get, string $operation): bool => $operation === 'create' || (bool) $get('change_password'))
                            ->visible(fn (Get $get, string $operation): bool => $operation === 'create' || (bool) $get('change_password'))
                            ->dehydrated(false)
                            ->maxLength(255)
                            ->helperText('Nhập lại đúng mật khẩu để tránh thao tác nhầm.'),
                    ])
                    ->columns(2),
                Section::make('Phân quyền')
                    ->schema([
                        $roleField,
                        Select::make('status')
                            ->label('Trạng thái')
                            ->options(EnumPresenter::options(UserStatus::class))
                            ->required(),
                    ])
                    ->columns(2),
                Section::make('Xác minh và audit')
                    ->schema([
                        DateTimePicker::make('phone_verified_at')
                            ->label('Xác minh số điện thoại')
                            ->disabled()
                            ->dehydrated(false),
                        DateTimePicker::make('last_login_at')
                            ->label('Đăng nhập cuối')
                            ->disabled()
                            ->dehydrated(false),
                    ])
                    ->columns(2),
                Section::make('Số dư ví')
                    ->description('Chỉnh trực tiếp số dư khả dụng của từng ví người dùng.')
                    ->schema([
                        TextInput::make('wallet_vnd_balance')
                            ->label('Số dư ví '.self::walletUnitLabel(UnitTransaction::VND))
                            ->numeric()
                            ->default(0)
                            ->step('0.000001'),
                        TextInput::make('wallet_usdt_balance')
                            ->label('Số dư ví '.self::walletUnitLabel(UnitTransaction::USDT))
                            ->numeric()
                            ->default(0)
                            ->step('0.000001'),
                    ])
                    ->columns(2),
                Section::make('Tài khoản rút tiền của khách hàng')
                    ->schema([
                        Section::make('Tài khoản ngân hàng thường')
                            ->schema([
                                TextInput::make('withdrawal_vnd_provider_code')
                                    ->label('Ngân hàng / Provider')
                                    ->maxLength(50),
                                TextInput::make('withdrawal_vnd_account_name')
                                    ->label('Chủ tài khoản')
                                    ->maxLength(255)
                                    ->requiredWith('withdrawal_vnd_account_number'),
                                TextInput::make('withdrawal_vnd_account_number')
                                    ->label('Số tài khoản')
                                    ->maxLength(255)
                                    ->requiredWith('withdrawal_vnd_account_name'),
                            ])
                            ->columns(1),
                        Section::make('Ví USDT')
                            ->schema([
                                TextInput::make('withdrawal_usdt_provider_code')
                                    ->label('Mạng / Provider')
                                    ->maxLength(50),
                                TextInput::make('withdrawal_usdt_account_name')
                                    ->label('Tên ví / Chủ sở hữu')
                                    ->maxLength(255)
                                    ->requiredWith('withdrawal_usdt_account_number'),
                                TextInput::make('withdrawal_usdt_account_number')
                                    ->label('Địa chỉ ví')
                                    ->maxLength(255)
                                    ->requiredWith('withdrawal_usdt_account_name'),
                            ])
                            ->columns(1),
                    ])
                    ->columns(2)
                    ->columnSpanFull(),
            ]);
    }

    private static function walletUnitLabel(UnitTransaction $unit): string
    {
        return match ($unit) {
            UnitTransaction::USDT => 'USDT',
            default => 'VND',
        };
    }
}
