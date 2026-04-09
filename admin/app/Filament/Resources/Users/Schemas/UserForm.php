<?php

namespace App\Filament\Resources\Users\Schemas;

use App\Enum\User\RoleUser;
use App\Enum\User\UserStatus;
use App\Support\Filament\EnumPresenter;
use Filament\Forms\Components\DateTimePicker;
use Filament\Forms\Components\Select;
use Filament\Forms\Components\TextInput;
use Filament\Schemas\Components\Section;
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
                        TextInput::make('email')
                            ->label('Email')
                            ->email()
                            ->required()
                            ->unique(ignoreRecord: true)
                            ->maxLength(255),
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
                        DateTimePicker::make('email_verified_at')
                            ->label('Xác minh email')
                            ->disabled(),
                        DateTimePicker::make('phone_verified_at')
                            ->label('Xác minh số điện thoại')
                            ->disabled(),
                        DateTimePicker::make('last_login_at')
                            ->label('Đăng nhập cuối')
                            ->disabled(),
                    ])
                    ->columns(3),
            ]);
    }
}
