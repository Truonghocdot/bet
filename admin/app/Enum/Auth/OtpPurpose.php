<?php

namespace App\Enum\Auth;

enum OtpPurpose: int
{
    case RESET_PASSWORD = 1;
    case VERIFY_EMAIL = 2;
    case VERIFY_PHONE = 3;
}

