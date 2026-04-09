<?php

namespace App\Enum\User;

enum UserStatus: int
{
    case ACTIVE = 1;
    case SUSPENDED = 2;
    case BANNED = 3;
}
