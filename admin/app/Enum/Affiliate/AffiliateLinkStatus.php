<?php

namespace App\Enum\Affiliate;

enum AffiliateLinkStatus: int
{
    case ACTIVE = 1;
    case PAUSED = 2;
}
