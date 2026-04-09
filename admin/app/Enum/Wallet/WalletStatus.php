<?php

namespace App\Enum\Wallet;

// Trạng thái của ví, ảnh hưởng đến việc có thể thực hiện giao dịch hay không 
enum WalletStatus: int
{
    case ACTIVE = 1;
    case LOCKED = 2;
}
