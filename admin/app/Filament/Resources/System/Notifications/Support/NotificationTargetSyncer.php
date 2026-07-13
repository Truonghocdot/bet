<?php

namespace App\Filament\Resources\System\Notifications\Support;

use App\Enum\Notification\NotificationAudience;
use App\Enum\Notification\NotificationResponseStatus;
use App\Models\Notification\Notification;

class NotificationTargetSyncer
{
    /**
     * @param  list<int|string>  $targetUserIds
     */
    public static function sync(Notification $notification, array $targetUserIds, bool $resetResponseStates = false): void
    {
        $audience = (int) ($notification->audience?->value ?? $notification->audience);
        if ($audience !== NotificationAudience::USERS->value) {
            $notification->targetUsers()->sync([]);

            return;
        }

        $ids = collect($targetUserIds)
            ->map(static fn (mixed $id): int => (int) $id)
            ->filter(static fn (int $id): bool => $id > 0)
            ->unique()
            ->values();

        if ($ids->isEmpty()) {
            $notification->targetUsers()->sync([]);

            return;
        }

        $hasResponseFlow = $notification->supportsResponseTracking();
        $currentIds = $notification->targetUsers()->pluck('users.id')->map(static fn (mixed $id): int => (int) $id)->all();
        $detachedIds = array_diff($currentIds, $ids->all());

        if ($detachedIds !== []) {
            $notification->targetUsers()->detach($detachedIds);
        }

        $syncPayload = [];
        foreach ($ids as $id) {
            $pivot = [];

            if (! $hasResponseFlow) {
                $pivot['response_status'] = null;
                $pivot['responded_at'] = null;
            } elseif ($resetResponseStates || ! in_array($id, $currentIds, true)) {
                $pivot['response_status'] = NotificationResponseStatus::PENDING->value;
                $pivot['responded_at'] = null;
            }

            $syncPayload[$id] = $pivot;
        }

        $notification->targetUsers()->syncWithoutDetaching($syncPayload);
    }
}
