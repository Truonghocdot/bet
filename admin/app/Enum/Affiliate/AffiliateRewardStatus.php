<?php

namespace App\Enum\Affiliate;

/**
 * Trạng thái chi trả thưởng theo mốc referral.
 */
enum AffiliateRewardStatus: int
{
    case PENDING = 1;
    case PAID = 2;
    case CANCELED = 3;
}
