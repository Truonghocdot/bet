<?php

namespace App\Enum\Bet;

/**
 * Nhóm cửa cược chuẩn hóa cho game round-based.
 *
 * option_key cụ thể sẽ phụ thuộc theo game_type.
 */
enum BetOptionType: int
{
    case NUMBER = 1; // cược số cụ thể
    case BIG_SMALL = 2; // lớn/nhỏ
    case ODD_EVEN = 3; // chẵn/lẻ
    case COLOR = 4; // màu (đỏ/xanh/tím...)
    case SUM = 5; // tổng
    case COMBINATION = 6; // tổ hợp
}
