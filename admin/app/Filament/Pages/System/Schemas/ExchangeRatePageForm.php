<?php

namespace App\Filament\Pages\System\Schemas;

use Filament\Forms\Components\DateTimePicker;
use Filament\Forms\Components\Textarea;
use Filament\Forms\Components\TextInput;
use Filament\Forms\Components\Toggle;
use Filament\Schemas\Components\Section;
use Filament\Schemas\Schema;

class ExchangeRatePageForm
{
    public static function configure(Schema $schema): Schema
    {
        return $schema->components([
            Section::make('Tỷ giá USDT/VND')
                ->description('Dữ liệu này được lưu vào Redis để service Gin và Laravel cùng tái sử dụng.')
                ->schema([
                    TextInput::make('code')
                        ->label('Mã cấu hình')
                        ->disabled(),
                    TextInput::make('base_currency')
                        ->label('Tiền gốc')
                        ->disabled(),
                    TextInput::make('quote_currency')
                        ->label('Tiền quy đổi')
                        ->disabled(),
                    TextInput::make('rate')
                        ->label('Tỷ giá áp dụng')
                        ->numeric()
                        ->required()
                        ->step('0.000001')
                        ->helperText('Giá đang dùng để quy đổi trong hệ thống.'),
                    TextInput::make('source_rate')
                        ->label('Tỷ giá nguồn')
                        ->disabled(),
                    Toggle::make('auto_sync')
                        ->label('Tự động đồng bộ từ nguồn'),
                    TextInput::make('source_name')
                        ->label('Nguồn cập nhật')
                        ->disabled(),
                    TextInput::make('redis_connection')
                        ->label('Kết nối Redis chia sẻ')
                        ->disabled()
                        ->helperText('Gin và Laravel đọc cùng một connection này.'),
                    TextInput::make('redis_key')
                        ->label('Redis key raw JSON')
                        ->disabled()
                        ->helperText('Gin đọc trực tiếp key này, không qua serializer Laravel.'),
                    TextInput::make('cache_store')
                        ->label('Cache store Laravel')
                        ->disabled(),
                    TextInput::make('cache_key')
                        ->label('Cache key Laravel')
                        ->disabled(),
                    DateTimePicker::make('last_synced_at')
                        ->label('Đồng bộ lần cuối')
                        ->disabled(),
                    Textarea::make('note')
                        ->label('Ghi chú')
                        ->rows(4)
                        ->columnSpanFull(),
                ])
                ->columns(2),

            Section::make('Cấu hình NOWPayments')
                ->description('Thông số kết nối cổng thanh toán Crypto.')
                ->schema([
                    TextInput::make('nowpayments_api_key')
                        ->label('NOWPayments API Key')
                        ->password()
                        ->revealable()
                        ->helperText('Lấy từ dashboard nowpayments.io'),
                    TextInput::make('nowpayments_ipn_secret')
                        ->label('IPN Secret Key')
                        ->password()
                        ->revealable()
                        ->helperText('Dùng để xác thực Webhook gửi từ NOWPayments.'),
                    TextInput::make('nowpayments_payout_wallet')
                        ->label('Ví nhận tiền (Payout Wallet)')
                        ->helperText('Địa chỉ ví USDT TRC20 để nhận tiền rút.'),
                    Toggle::make('nowpayments_sandbox')
                        ->label('Chế độ Sandbox (Test)')
                        ->helperText('Bật nếu đang dùng môi trường thử nghiệm.'),
                ])
                ->columns(2),

            Section::make('Khối Thông Tin Rút Tiền')
                ->description('Chỉ hiển thị trên app người chơi, không còn áp dụng làm rule cứng ở backend Gin.')
                ->schema([
                    Toggle::make('withdraw_policy_enabled')
                        ->label('Hiển thị khối thông tin ở màn rút tiền'),
                    TextInput::make('withdraw_fee_percent')
                        ->label('Lệ phí (%)')
                        ->numeric()
                        ->minValue(0)
                        ->step('0.01')
                        ->required()
                        ->helperText('Chỉ dùng để hiển thị trên app. Ví dụ: 0, 1.5, 2 ...'),
                    TextInput::make('withdraw_required_bet_volume')
                        ->label('Tổng tiền cược')
                        ->numeric()
                        ->minValue(0)
                        ->step('0.000001')
                        ->required()
                        ->helperText('Chỉ dùng để hiển thị trên app.'),
                    TextInput::make('withdraw_max_times_per_day')
                        ->label('Số lần rút tiền')
                        ->numeric()
                        ->minValue(1)
                        ->step('1')
                        ->required()
                        ->helperText('Chỉ dùng để hiển thị trên app.'),
                    TextInput::make('withdraw_min_amount')
                        ->label('Rút tối thiểu')
                        ->numeric()
                        ->minValue(0)
                        ->step('0.000001')
                        ->required()
                        ->helperText('Chỉ dùng để hiển thị trên app.'),
                    TextInput::make('withdraw_max_amount')
                        ->label('Rút tối đa')
                        ->numeric()
                        ->minValue(0)
                        ->step('0.000001')
                        ->required()
                        ->helperText('Chỉ dùng để hiển thị trên app.'),
                ])
                ->columns(2),

            Section::make('Cấu hình Hỗ trợ & Mạng xã hội')
                ->description('Thông tin liên hệ CSKH hiển thị trên ứng dụng.')
                ->schema([
                    TextInput::make('telegram_cskh_link')
                        ->label('Link Telegram CSKH')
                        ->url()
                        ->placeholder('https://t.me/your_cskh')
                        ->helperText('Đường dẫn trực tiếp đến tài khoản hoặc group Telegram hỗ trợ.'),
                ])
                ->columns(2),

            Section::make('Marquee thông báo')
                ->description('Nội dung chạy ngang ở trang chủ app. Mỗi dòng là một thông báo riêng.')
                ->schema([
                    Toggle::make('marquee_enabled')
                        ->label('Bật marquee trên app'),
                    Textarea::make('marquee_messages')
                        ->label('Danh sách thông báo')
                        ->rows(6)
                        ->columnSpanFull()
                        ->placeholder("Quý khách thân mến vui lòng thay đổi cổng nạp tiền nếu không thể tạo lệnh nạp.\nKhi nạp tiền bằng cổng CHUYỂN KHOẢN sẽ được nhận thêm ưu đãi đặc biệt!\nFF789 - Đăng ký hôm nay nhận ngay thưởng chào mừng 100%.")
                        ->helperText('Mỗi dòng tương ứng một câu chạy trong marquee.'),
                ])
                ->columns(2),

            Section::make('Popup thông báo')
                ->description('Cấu hình 2 popup riêng biệt hiển thị trên app.')
                ->schema([
                    Textarea::make('popup_message')
                        ->label('Popup thông báo')
                        ->rows(8)
                        ->placeholder("Chào mừng bạn đến với FF789.\nLiên hệ CSKH nếu cần hỗ trợ đổi cổng nạp.")
                        ->helperText('Nội dung popup thông báo chung. Để trống nếu không hiển thị.'),
                    Textarea::make('latest_news_popup')
                        ->label('Tin tức popup mới nhất')
                        ->rows(8)
                        ->placeholder("Sự kiện hoàn trả cuối tuần đang diễn ra.\nCập nhật ưu đãi mới nhất tại FF789.")
                        ->helperText('Nội dung popup dành cho tin tức mới nhất. Để trống nếu không hiển thị.'),
                ])
                ->columns(2),
        ]);
    }
}
