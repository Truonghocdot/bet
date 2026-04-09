<?php

namespace App\Enum\Transaction;

enum TransactionStatus: int
{
    case PENDING = 1;
    case CONFIRMED = 2;
    case COMPLETED = 3;
    case FAILED = 4;
    case CANCELED = 5;
}
