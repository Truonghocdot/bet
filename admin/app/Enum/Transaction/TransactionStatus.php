<?php

namespace App\Enum\Transaction;


// Trạng thái giao dịch
enum TransactionStatus: int
{
    case PENDING = 1;
    case CONFIRMED = 2;
    case COMPLETED = 3;
    case FAILED = 4;
    // Với lệnh nạp: có thể do hệ thống auto-cancel khi hết hạn hoặc do người dùng chủ động hủy.
    case CANCELED = 5;
}
