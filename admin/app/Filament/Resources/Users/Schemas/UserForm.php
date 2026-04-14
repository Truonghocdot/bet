<?php

namespace App\Filament\Resources\Users\Schemas;

use App\Enum\Affiliate\AffiliateProfileStatus;
use App\Enum\User\RoleUser;
use App\Enum\User\UserStatus;
use App\Enum\Wallet\UnitTransaction;
use App\Support\Filament\EnumPresenter;
use Filament\Forms\Components\DateTimePicker;
use Filament\Forms\Components\Toggle;
use Filament\Forms\Components\Select;
use Filament\Forms\Components\TextInput;
use Filament\Schemas\Components\Section;
use Filament\Schemas\Components\Utilities\Get;
use Filament\Schemas\Schema;

class UserForm
{
    public static function configure(Schema $schema): Schema
    {
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
                        Select::make('role')
                            ->label('Vai trò')
                            ->options(EnumPresenter::options(RoleUser::class))
                            ->required(),
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
                            ->disabled(),
                        DateTimePicker::make('last_login_at')
                            ->label('Đăng nhập cuối')
                            ->disabled(),
                    ])
                    ->columns(3),
                Section::make('Khởi tạo dữ liệu liên quan')
                    ->visible(fn (string $operation): bool => $operation === 'create')
                    ->schema([
                        Toggle::make('provision_wallets')
                            ->label('Tạo ví mặc định')
                            ->default(true)
                            ->live()
                            ->helperText('Khuyến nghị bật để user có thể nạp/rút/cược ngay sau khi tạo.'),
                        Toggle::make('provision_vnd_wallet')
                            ->label('Tạo ví VND')
                            ->default(true)
                            ->visible(fn (Get $get): bool => (bool) $get('provision_wallets'))
                            ->helperText('Ví tiền chính để nạp/rút và cược.'),
                        Toggle::make('provision_affiliate_profile')
                            ->label('Tạo hồ sơ affiliate')
                            ->default(false)
                            ->live()
                            ->helperText('Bật nếu user này cần quản lý referral ngay từ đầu. Mã giới thiệu sẽ tự sinh duy nhất.'),
                        Select::make('affiliate_status')
                            ->label('Trạng thái affiliate')
                            ->options(EnumPresenter::options(AffiliateProfileStatus::class))
                            ->default(AffiliateProfileStatus::PENDING->value)
                            ->visible(fn (Get $get): bool => (bool) $get('provision_affiliate_profile')),
                        Toggle::make('provision_account_withdrawal_info')
                            ->label('Tạo tài khoản rút')
                            ->default(false)
                            ->live()
                            ->helperText('Bật nếu muốn lưu sẵn tài khoản nhận tiền cho user.'),
                        Select::make('withdrawal_unit')
                            ->label('Đơn vị tài khoản rút')
                            ->options(EnumPresenter::options(UnitTransaction::class))
                            ->default(UnitTransaction::VND->value)
                            ->required(fn (Get $get): bool => (bool) $get('provision_account_withdrawal_info'))
                            ->visible(fn (Get $get): bool => (bool) $get('provision_account_withdrawal_info')),
                        TextInput::make('withdrawal_provider_code')
                            ->label('Mã nhà cung cấp')
                            ->maxLength(50)
                            ->required(fn (Get $get): bool => (bool) $get('provision_account_withdrawal_info'))
                            ->visible(fn (Get $get): bool => (bool) $get('provision_account_withdrawal_info')),
                        TextInput::make('withdrawal_account_name')
                            ->label('Chủ tài khoản')
                            ->maxLength(255)
                            ->required(fn (Get $get): bool => (bool) $get('provision_account_withdrawal_info'))
                            ->visible(fn (Get $get): bool => (bool) $get('provision_account_withdrawal_info')),
                        TextInput::make('withdrawal_account_number')
                            ->label('Số tài khoản')
                            ->maxLength(255)
                            ->required(fn (Get $get): bool => (bool) $get('provision_account_withdrawal_info'))
                            ->visible(fn (Get $get): bool => (bool) $get('provision_account_withdrawal_info')),
                        Toggle::make('withdrawal_is_default')
                            ->label('Mặc định')
                            ->visible(fn (Get $get): bool => (bool) $get('provision_account_withdrawal_info')),
                    ])
                    ->columns(2),
            ]);
    }
}
