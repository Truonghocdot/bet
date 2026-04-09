<?php

namespace App\Enum\Wallet;

enum WalletStatus: int
{
    case ACTIVE = 1;
    case LOCKED = 2;
}
