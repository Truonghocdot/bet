<?php

namespace App\Enum\Bet;

/**
 * Kết quả chấm cho từng dòng cược trong ticket.
 */
enum BetItemResult: int
{
    case PENDING = 1; // chưa chấm
    case WON = 2; // thắng
    case LOST = 3; // thua
    case VOID = 4; // hoàn cược
    case HALF_WON = 5; // thắng nửa
    case HALF_LOST = 6; // thua nửa
}
