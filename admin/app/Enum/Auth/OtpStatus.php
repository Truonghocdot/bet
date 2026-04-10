<?php

namespace App\Enum\Auth;

enum OtpStatus: int
{
    case PENDING = 1;
    case VERIFIED = 2;
    case USED = 3;
    case EXPIRED = 4;
    case LOCKED = 5;
    case CANCELLED = 6;
}

