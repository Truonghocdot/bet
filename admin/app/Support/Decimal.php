<?php

namespace App\Support;

final class Decimal
{
    public const int SCALE = 8;

    public const string ZERO = '0.00000000';

    public static function normalize(mixed $value, bool $signed = false): ?string
    {
        if (! is_string($value) && ! is_int($value)) {
            return null;
        }

        $normalized = str_replace([',', ' '], ['', ''], trim((string) $value));
        $pattern = $signed
            ? '/\A[+-]?\d+(?:\.\d{1,8})?\z/D'
            : '/\A\+?\d+(?:\.\d{1,8})?\z/D';

        if ($normalized === '' || preg_match($pattern, $normalized) !== 1) {
            return null;
        }

        return bcadd($normalized, self::ZERO, self::SCALE);
    }
}
