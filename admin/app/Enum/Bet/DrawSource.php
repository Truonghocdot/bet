<?php

namespace App\Enum\Bet;

/**
 * Nguồn dữ liệu kết quả quay số.
 *
 * Dùng để audit và kiểm soát độ tin cậy của kết quả.
 */
enum DrawSource: int
{
    case AUTO = 1; // lấy tự động từ provider/hệ thống
    case MANUAL = 2; // nhập tay bởi admin
    case IMPORTED = 3; // import batch
}
