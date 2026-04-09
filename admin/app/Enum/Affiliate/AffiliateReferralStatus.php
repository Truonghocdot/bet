<?php

namespace App\Enum\Affiliate;

/**
 * Trạng thái người được mời.
 */
enum AffiliateReferralStatus: int
{
    case PENDING = 1; // đã gắn ref, chưa đạt điều kiện nạp tối thiểu
    case QUALIFIED = 2; // đã nạp đủ điều kiện (>= 50.000 VND)
    case INVALID = 3; // không hợp lệ (tài khoản gian lận, tự giới thiệu...)
}
