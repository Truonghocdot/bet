<?php

namespace App\Enum\Transaction;

enum TypeTransaction: int
{
    case DEPOSIT = 1;
    case WITHDRAW = 2;
}
