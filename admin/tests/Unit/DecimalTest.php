<?php

namespace Tests\Unit;

use App\Support\Decimal;
use PHPUnit\Framework\TestCase;

class DecimalTest extends TestCase
{
    public function test_it_normalizes_decimal_strings_without_float_precision_loss(): void
    {
        self::assertSame(
            '9999999999999999999999.12345678',
            Decimal::normalize('9,999,999,999,999,999,999,999.12345678'),
        );
        self::assertSame('12.34000000', Decimal::normalize('00012.3400'));
        self::assertSame('1000.00000000', Decimal::normalize(1000));
        self::assertSame('-12.34000000', Decimal::normalize('-12.34', signed: true));
        self::assertSame('12.34000000', Decimal::normalize('+12.34', signed: true));
    }

    public function test_it_rejects_values_that_cannot_be_stored_exactly_at_scale_eight(): void
    {
        self::assertNull(Decimal::normalize('1.123456789'));
        self::assertNull(Decimal::normalize('-1'));
        self::assertNull(Decimal::normalize('1e3'));
        self::assertNull(Decimal::normalize(1.25));
        self::assertNull(Decimal::normalize('not-a-number'));
    }
}
