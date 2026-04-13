<?php

namespace App\Enum\Notification;

enum NotificationStatus: int
{
    case DRAFT = 1;
    case PUBLISHED = 2;
    case ARCHIVED = 3;
}

