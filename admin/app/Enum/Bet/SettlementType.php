<?php

namespace App\Enum\Bet;

/**
 * Cách thực hiện settlement.
 *
 * AUTO: job tự động chấm.
 * MANUAL: admin chấm tay/override.
 * ROLLBACK: hoàn tác settlement trước đó để chấm lại.
 */
enum SettlementType: int
{
    case AUTO = 1;
    case MANUAL = 2;
    case ROLLBACK = 3;
}
