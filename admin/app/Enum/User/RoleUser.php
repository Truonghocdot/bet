<?php

namespace App\Enum\User;

enum RoleUser: int // vai trò người dùng để phân quyền truy cập vào hệ thống
{
    case SUPER_ADMIN = 0;
    case ADMIN = 1;
    case CLIENT = 2;
    case STAFF = 3;
    case AGENCY = 4;

    /**
     * @return array<int, self>
     */
    public static function manageableBy(?self $actorRole): array
    {
        return match ($actorRole) {
            self::SUPER_ADMIN => [self::ADMIN, self::STAFF, self::CLIENT, self::AGENCY],
            self::ADMIN => [self::STAFF, self::CLIENT, self::AGENCY],
            default => [],
        };
    }

    /**
     * @return array<int, int>
     */
    public static function manageableValuesBy(?self $actorRole): array
    {
        return array_map(
            static fn (self $role): int => $role->value,
            self::manageableBy($actorRole),
        );
    }

    public static function canAssign(?self $actorRole, self $targetRole): bool
    {
        return in_array($targetRole, self::manageableBy($actorRole), true);
    }
}
