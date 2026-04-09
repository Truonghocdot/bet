<?php

namespace App\Enum\Transaction;

enum WithdrawalStatus: int
{
    case PENDING = 1;
    case APPROVED = 2;
    case REJECTED = 3;
    case CANCELED = 4;
    case PAID = 5;
}
