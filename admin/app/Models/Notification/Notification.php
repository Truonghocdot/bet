<?php

namespace App\Models\Notification;

use App\Enum\Notification\NotificationAudience;
use App\Enum\Notification\NotificationResponseStatus;
use App\Enum\Notification\NotificationStatus;
use App\Models\User;
use App\Support\Media\WebpImageConverter;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\Relations\BelongsToMany;

class Notification extends Model
{
    protected $fillable = [
        'title',
        'body',
        'image_path',
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

    protected static function booted(): void
    {
        static::saving(function (self $notification): void {
            $notification->image_path = WebpImageConverter::convertPublicDiskPath($notification->image_path);
        });
    }

    public function createdBy(): BelongsTo
    {
        return $this->belongsTo(User::class, 'created_by');
    }

    public function targetUsers(): BelongsToMany
    {
        return $this->belongsToMany(User::class, 'notification_targets', 'notification_id', 'user_id')
            ->withPivot('response_status', 'responded_at');
    }

    public function reads(): BelongsToMany
    {
        return $this->belongsToMany(User::class, 'notification_reads', 'notification_id', 'user_id')
            ->withPivot('read_at');
    }

    public function pendingResponseTargets(): BelongsToMany
    {
        return $this->targetUsers()->wherePivot('response_status', NotificationResponseStatus::PENDING->value);
    }

    public function confirmedResponseTargets(): BelongsToMany
    {
        return $this->targetUsers()->wherePivot('response_status', NotificationResponseStatus::CONFIRMED->value);
    }

    public function canceledResponseTargets(): BelongsToMany
    {
        return $this->targetUsers()->wherePivot('response_status', NotificationResponseStatus::CANCELED->value);
    }

    public function supportsResponseTracking(): bool
    {
        return filled($this->image_path)
            && (int) ($this->audience?->value ?? $this->audience) === NotificationAudience::USERS->value;
    }
}
