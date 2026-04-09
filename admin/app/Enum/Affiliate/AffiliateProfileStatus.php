<?php

namespace App\Enum\Affiliate;

enum AffiliateProfileStatus: int
{
    case PENDING = 1;
    case ACTIVE = 2;
    case SUSPENDED = 3;
}
