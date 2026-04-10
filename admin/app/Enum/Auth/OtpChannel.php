<?php

namespace App\Enum\Auth;

enum OtpChannel: int
{
    case EMAIL = 1;
    case PHONE = 2;
}

