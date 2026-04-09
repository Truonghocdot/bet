<?php

namespace App\Enum\Bet;

/**
 * Trạng thái ticket cược.
 *
 * Trạng thái terminal:
 * WON, LOST, VOID, HALF_WON, HALF_LOST, CANCELED, CASHED_OUT.
 */
enum BetStatus: int
{
    case PENDING = 1; // chưa chấm
    case WON = 2; // thắng
    case LOST = 3; // thua
    case VOID = 4; // hoàn cược
    case HALF_WON = 5; // thắng nửa
    case HALF_LOST = 6; // thua nửa
    case CANCELED = 7; // hủy vé
    case CASHED_OUT = 8; // tất toán sớm
}
