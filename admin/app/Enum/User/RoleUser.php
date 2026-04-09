<?php

namespace App\Enum\User;

enum RoleUser: int // vai trò người dùng để phân quyền truy cập vào hệ thống
{
    case ADMIN = 1;
    case CLIENT = 2;
    case STAFF = 3;
    case AGENCY = 4;
}