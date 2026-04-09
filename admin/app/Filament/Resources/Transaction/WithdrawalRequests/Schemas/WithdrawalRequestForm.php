<?php

namespace App\Filament\Resources\Transaction\WithdrawalRequests\Schemas;

use App\Enum\Transaction\WithdrawalStatus;
use App\Enum\Wallet\UnitTransaction;
use App\Support\Filament\EnumPresenter;
use Filament\Forms\Components\DateTimePicker;
use Filament\Forms\Components\FileUpload;
use Filament\Schemas\Components\Section;
use Filament\Forms\Components\Select;
use Filament\Forms\Components\TextInput;
use Filament\Forms\Components\Textarea;
use Filament\Schemas\Schema;

class WithdrawalRequestForm
{
    public static function configure(Schema $schema): Schema
    {
        return $schema->components([
            Section::make('Thông tin yêu cầu rút')
                ->schema([
                    Select::make('user_id')
                        ->label('Người dùng')
                        ->relationship('user', 'name')
                        ->searchable()
                        ->preload()
                        ->required(),
                    Select::make('wallet_id')
                        ->label('Ví')
                        ->relationship('wallet', 'id')
                        ->searchable()
                        ->preload()
                        ->required(),
                    Select::make('account_withdrawal_info_id')
                        ->label('Tài khoản rút')
                        ->relationship('accountWithdrawalInfo', 'account_number')
                        ->searchable()
                        ->preload()
                        ->required(),
                    Select::make('unit')
                        ->label('Đơn vị')
                        ->options(EnumPresenter::options(UnitTransaction::class))
                        ->required(),
                    TextInput::make('amount')
                        ->label('Số tiền')
                        ->numeric()
                        ->required(),
                    TextInput::make('fee')
                        ->label('Phí')
                        ->numeric()
                        ->required()
                        ->default(0),
                    TextInput::make('net_amount')
                        ->label('Thực nhận')
                        ->numeric()
                        ->required(),
                    Select::make('status')
                        ->label('Trạng thái')
                        ->options(EnumPresenter::options(WithdrawalStatus::class))
                        ->required(),
                    Textarea::make('reason_rejected')
                        ->label('Lý do từ chối')
                        ->rows(3)
                        ->columnSpanFull(),
                    Select::make('reviewed_by')
                        ->label('Duyệt boi')
                        ->relationship('reviewedBy', 'name')
                        ->searchable()
                        ->preload(),
                    DateTimePicker::make('reviewed_at')
                        ->label('Duyệt lúc'),
                    Select::make('paid_by')
                        ->label('Chi trả bởi')
                        ->relationship('paidBy', 'name')
                        ->searchable()
                        ->preload(),
                    DateTimePicker::make('paid_at')
                        ->label('Chi trả lúc'),
                    TextInput::make('transfer_reference')
                        ->label('Mã giao dịch/chứng từ')
                        ->maxLength(255),
                    FileUpload::make('transfer_proof_path')
                        ->label('Ảnh chứng từ')
                        ->disk('public')
                        ->directory('withdrawal-proofs')
                        ->preserveFilenames(),
                    Textarea::make('admin_note')
                        ->label('Ghi chú admin')
                        ->rows(3)
                        ->columnSpanFull(),
                ])
                ->columns(2),
        ]);
    }
}
