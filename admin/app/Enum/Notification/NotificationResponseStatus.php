<?php

namespace App\Enum\Notification;

enum NotificationResponseStatus: int
{
    case PENDING = 1;
    case CONFIRMED = 2;
    case CANCELED = 3;
}
