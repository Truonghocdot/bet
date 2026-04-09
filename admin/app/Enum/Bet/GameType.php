<?php

namespace App\Enum\Bet;

/**
 * Loại game round-based.
 *
 * Dùng để route business logic:
 * - parser kết quả
 * - rule cược hợp lệ
 * - công thức payout
 */
enum GameType: int
{
    case WINGO = 1;
    case K3 = 2;
    case LOTTERY = 3;
}
