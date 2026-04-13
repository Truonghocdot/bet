<?php

namespace App\Models\Notification;

use App\Enum\Notification\NotificationAudience;
use App\Enum\Notification\NotificationStatus;
use App\Models\User;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\Relations\BelongsToMany;

class Notification extends Model
{
    protected $fillable = [
        'title',
        'body',
        'status',
        'audience',
        'publish_at',
        'expires_at',
        'created_by',
    ];

    protected function casts(): array
    {
        return [
            'status' => NotificationStatus::class,
            'audience' => NotificationAudience::class,
            'publish_at' => 'datetime',
            'expires_at' => 'datetime',
        ];
    }

    public function createdBy(): BelongsTo
    {
        return $this->belongsTo(User::class, 'created_by');
    }

    public function targetUsers(): BelongsToMany
    {
        return $this->belongsToMany(User::class, 'notification_targets', 'notification_id', 'user_id');
    }

    public function reads(): BelongsToMany
    {
        return $this->belongsToMany(User::class, 'notification_reads', 'notification_id', 'user_id')
            ->withPivot('read_at');
    }
}

