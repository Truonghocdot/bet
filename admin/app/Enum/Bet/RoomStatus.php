<?php

namespace App\Enum\Bet;

/**
 * Trạng thái room game.
 */
enum RoomStatus: int
{
    case ACTIVE = 1;
    case INACTIVE = 2;
}
