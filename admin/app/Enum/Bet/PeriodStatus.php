<?php

namespace App\Enum\Bet;

/**
 * Trạng thái vòng đời của 1 kỳ cược.
 *
 * Transition chuẩn:
 * SCHEDULED -> OPEN -> LOCKED -> DRAWN -> SETTLED
 * và có thể về CANCELED trước SETTLED.
 */
enum PeriodStatus: int
{
    case SCHEDULED = 1; // đã tạo kỳ, chưa mở cược
    case OPEN = 2; // đang mở nhận cược
    case LOCKED = 3; // đã khóa cược, chờ kết quả
    case DRAWN = 4; // đã có kết quả raw, chờ settlement
    case SETTLED = 5; // đã chấm và cập nhật ví
    case CANCELED = 6; // kỳ bị hủy
}
