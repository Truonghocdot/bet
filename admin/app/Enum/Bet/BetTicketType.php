<?php

namespace App\Enum\Bet;

/**
 * Loại vé cược.
 *
 * SINGLE: 1 lựa chọn.
 * MULTI: nhiều lựa chọn trong 1 ticket.
 */
enum BetTicketType: int
{
    case SINGLE = 1;
    case MULTI = 2;
}
