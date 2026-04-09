<?php

namespace App\Enum\Transaction;

// Loại giao dịch
enum TypeTransaction: int
{
    case DEPOSIT = 1;
    case WITHDRAW = 2;
}
