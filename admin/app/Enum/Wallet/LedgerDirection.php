<?php

namespace App\Enum\Wallet;

enum LedgerDirection: int
{
    case CREDIT = 1;
    case DEBIT = 2;
}
