<?php

namespace App\Enum\Payment;

/**
 * Trạng thái hiển thị của tài khoản nhận tiền.
 */
enum PaymentReceivingAccountStatus: int
{
    case ACTIVE = 1;
    case INACTIVE = 2;
}
