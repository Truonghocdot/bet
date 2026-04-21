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
use Filament\Schemas\Components\Section;
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
                        TextInput::make('password')
                            ->label('Mật khẩu')
                            ->password()
                            ->revealable()
                            ->required(fn (string $operation): bool => $operation === 'create')
                            ->dehydrated(fn (?string $state): bool => filled($state))
                            ->maxLength(255)
                            ->helperText('Để trống khi sửa nếu không muốn đổi mật khẩu.'),
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
                Section::make('Tài khoản rút mặc định')
                    ->description('Cho phép cập nhật nhanh tài khoản rút chính ngay trong hồ sơ người dùng.')
                    ->schema([
                        Select::make('withdrawal_unit')
                            ->label('Đơn vị')
                            ->options(EnumPresenter::options(UnitTransaction::class))
                            ->default(UnitTransaction::VND->value),
                        TextInput::make('withdrawal_provider_code')
                            ->label('Ngân hàng / Provider')
                            ->maxLength(50),
                        TextInput::make('withdrawal_account_name')
                            ->label('Chủ tài khoản')
                            ->maxLength(255)
                            ->requiredWith('withdrawal_account_number'),
                        TextInput::make('withdrawal_account_number')
                            ->label('Số tài khoản')
                            ->maxLength(255)
                            ->requiredWith('withdrawal_account_name'),
                        Hidden::make('withdrawal_is_default')
                            ->default(true),
                    ])
                    ->columns(2),
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
