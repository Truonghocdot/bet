<?php

namespace App\Enum\Wallet;

// Hướng của bút toán trong sổ cái ví, ảnh hưởng đến việc tăng hay giảm số dư
enum LedgerDirection: int
{
    case CREDIT = 1;
    case DEBIT = 2;
}
