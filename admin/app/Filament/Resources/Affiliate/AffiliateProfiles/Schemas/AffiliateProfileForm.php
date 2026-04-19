<?php

namespace App\Filament\Resources\Affiliate\AffiliateProfiles\Schemas;

use App\Enum\Affiliate\AffiliateProfileStatus;
use App\Enum\Affiliate\AffiliateReferralStatus;
use App\Enum\User\RoleUser;
use App\Models\Affiliate\AffiliateProfile;
use App\Support\Filament\EnumPresenter;
use Filament\Forms\Components\Placeholder;
use Filament\Forms\Components\Select;
use Filament\Forms\Components\TextInput;
use Filament\Schemas\Components\Section;
use Filament\Schemas\Schema;
use Illuminate\Support\HtmlString;

class AffiliateProfileForm
{
    public static function configure(Schema $schema): Schema
    {
        return $schema
            ->components([
                Section::make('Hồ sơ affiliate')
                    ->schema([
                        Select::make('user_id')
                            ->label('Người dùng')
                            ->relationship('user', 'name')
                            ->searchable()
                            ->preload()
                            ->required(),
                        TextInput::make('ref_code')
                            ->label('Mã giới thiệu')
                            ->disabled()
                            ->dehydrated(false)
                            ->helperText('Mã giới thiệu được hệ thống tự sinh và không nhập tay.')
                            ->maxLength(50),
                        TextInput::make('ref_link')
                            ->label('Link giới thiệu')
                            ->disabled()
                            ->dehydrated(false)
                            ->helperText('Link giới thiệu được tạo tự động từ mã giới thiệu.')
                            ->maxLength(255),
                        Select::make('status')
                            ->label('Trạng thái')
                            ->options(EnumPresenter::options(AffiliateProfileStatus::class))
                            ->required(),
                    ])
                    ->columns(2),
                Section::make('Tra cứu khách trực thuộc')
                    ->description('Dùng để kiểm tra khách nạp tiền đang thuộc tuyến đại lý/affiliate nào.')
                    ->schema([
                        Placeholder::make('affiliate_owner_summary')
                            ->label('Đại lý / chủ tuyến')
                            ->content(fn (?AffiliateProfile $record): HtmlString => new HtmlString(self::buildOwnerSummary($record))),
                        Placeholder::make('affiliate_referral_stats')
                            ->label('Thống kê khách trực thuộc')
                            ->content(fn (?AffiliateProfile $record): HtmlString => new HtmlString(self::buildReferralStats($record))),
                        Placeholder::make('affiliate_recent_referrals')
                            ->label('Khách gần nhất trong tuyến')
                            ->content(fn (?AffiliateProfile $record): HtmlString => new HtmlString(self::buildReferralPreview($record)))
                            ->columnSpanFull(),
                    ])
                    ->columns(2)
                    ->visible(fn (?AffiliateProfile $record): bool => filled($record?->getKey())),
            ]);
    }

    private static function buildOwnerSummary(?AffiliateProfile $record): string
    {
        if (! $record?->exists || ! $record->user) {
            return '<span class="text-gray-500">Lưu hồ sơ affiliate trước để hiển thị thông tin tuyến.</span>';
        }

        $owner = $record->user;
        $roleLabel = EnumPresenter::label(RoleUser::class, $owner->role);
        $phone = filled($owner->phone) ? e((string) $owner->phone) : '<span class="text-gray-500">Chưa có SĐT</span>';

        return sprintf(
            '<div><strong>%s</strong></div><div>Vai trò: %s</div><div>SĐT: %s</div><div>Mã giới thiệu: <strong>%s</strong></div>',
            e((string) $owner->name),
            e($roleLabel),
            $phone,
            e((string) $record->ref_code)
        );
    }

    private static function buildReferralStats(?AffiliateProfile $record): string
    {
        if (! $record?->exists) {
            return '<span class="text-gray-500">Chưa có dữ liệu khách trực thuộc.</span>';
        }

        $referrals = $record->referrals();
        $total = (int) $referrals->count();
        $withFirstDeposit = (int) $record->referrals()->whereNotNull('first_deposit_transaction_id')->count();
        $qualified = (int) $record->referrals()->whereNotNull('qualified_at')->count();
        $firstDepositTotal = (float) $record->referrals()->whereNotNull('first_deposit_transaction_id')->sum('first_deposit_amount');

        return sprintf(
            '<div>Tổng khách trực thuộc: <strong>%s</strong></div><div>Khách đã nạp đầu: <strong>%s</strong></div><div>Khách đủ điều kiện: <strong>%s</strong></div><div>Tổng nạp đầu ghi nhận: <strong>%sđ</strong></div>',
            number_format($total, 0, ',', '.'),
            number_format($withFirstDeposit, 0, ',', '.'),
            number_format($qualified, 0, ',', '.'),
            number_format($firstDepositTotal, 0, ',', '.')
        );
    }

    private static function buildReferralPreview(?AffiliateProfile $record): string
    {
        if (! $record?->exists) {
            return '<span class="text-gray-500">Chưa có khách trực thuộc.</span>';
        }

        $referrals = $record->referrals()
            ->with([
                'referredUser:id,name,phone',
                'firstDepositTransaction:id,client_ref',
            ])
            ->latest('created_at')
            ->limit(10)
            ->get();

        if ($referrals->isEmpty()) {
            return '<span class="text-gray-500">Affiliate này chưa có khách nào trong tuyến.</span>';
        }

        $rows = $referrals->map(function ($referral): string {
            $customerName = $referral->referredUser?->name ?: 'Khách không xác định';
            $customerPhone = $referral->referredUser?->phone ?: '---';
            $status = EnumPresenter::label(AffiliateReferralStatus::class, $referral->status);
            $depositAmount = $referral->first_deposit_amount !== null
                ? number_format((float) $referral->first_deposit_amount, 0, ',', '.').'đ'
                : 'Chưa nạp đầu';
            $transactionNo = $referral->firstDepositTransaction?->client_ref ?: '---';

            return sprintf(
                '<tr>
                    <td style="padding:8px;border-bottom:1px solid #e5e7eb;"><strong>%s</strong><br><span style="color:#6b7280">%s</span></td>
                    <td style="padding:8px;border-bottom:1px solid #e5e7eb;">%s</td>
                    <td style="padding:8px;border-bottom:1px solid #e5e7eb;">%s</td>
                    <td style="padding:8px;border-bottom:1px solid #e5e7eb;">%s</td>
                </tr>',
                e((string) $customerName),
                e((string) $customerPhone),
                e($status),
                e($depositAmount),
                e((string) $transactionNo),
            );
        })->implode('');

        return sprintf(
            '<div style="margin-bottom:8px;color:#6b7280;">Khi khách vừa nạp, có thể dò nhanh trong danh sách này để biết đang thuộc tuyến đại lý nào.</div>
            <div style="overflow:auto;">
                <table style="width:100%%;border-collapse:collapse;font-size:12px;">
                    <thead>
                        <tr style="background:#f8fafc;text-align:left;">
                            <th style="padding:8px;border-bottom:1px solid #e5e7eb;">Khách</th>
                            <th style="padding:8px;border-bottom:1px solid #e5e7eb;">Trạng thái</th>
                            <th style="padding:8px;border-bottom:1px solid #e5e7eb;">Nạp đầu</th>
                            <th style="padding:8px;border-bottom:1px solid #e5e7eb;">Mã GD nạp đầu</th>
                        </tr>
                    </thead>
                    <tbody>%s</tbody>
                </table>
            </div>',
            $rows
        );
    }
}
