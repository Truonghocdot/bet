<?php

namespace App\Enum\Wallet; 

enum UnitTransaction: int // đơn vị tiền tệ giao dịch nạp / rút (hiện tại có 2 đơn vị nạp/rút là VND và USDT)
{
    case VND = 1;
    case USDT = 2;
}