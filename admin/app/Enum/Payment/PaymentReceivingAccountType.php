<?php

namespace App\Enum\Payment;

/**
 * Loại tài khoản nhận tiền.
 *
 * BANK: tài khoản ngân hàng.
 * CRYPTO: ví tiền ảo / blockchain address.
 */
enum PaymentReceivingAccountType: int
{
    case BANK = 1;
    case CRYPTO = 2;
}
