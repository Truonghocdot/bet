<?php

namespace App\Filament\Resources\System\Notifications\Pages;

use App\Enum\Notification\NotificationAudience;
use App\Enum\Notification\NotificationStatus;
use App\Filament\Resources\System\Notifications\NotificationResource;
use Filament\Resources\Pages\EditRecord;
use Illuminate\Validation\ValidationException;

class EditNotification extends EditRecord
{
    protected static string $resource = NotificationResource::class;

    protected function mutateFormDataBeforeSave(array $data): array
    {
        $status = (int) ($data['status'] ?? 0);
        $audience = (int) ($data['audience'] ?? 0);

        if ($status === NotificationStatus::PUBLISHED->value && blank($data['publish_at'] ?? null)) {
            $data['publish_at'] = now();
        }

        if (
            filled($data['publish_at'] ?? null)
            && filled($data['expires_at'] ?? null)
            && strtotime((string) $data['expires_at']) <= strtotime((string) $data['publish_at'])
        ) {
            throw ValidationException::withMessages([
                'expires_at' => 'Thời gian hết hạn phải lớn hơn thời gian phát hành.',
            ]);
        }

        $targetUsers = $data['targetUsers'] ?? ($this->data['targetUsers'] ?? []);
        if ($audience === NotificationAudience::USERS->value && empty($targetUsers)) {
            throw ValidationException::withMessages([
                'targetUsers' => 'Vui lòng chọn ít nhất 1 người dùng đích.',
            ]);
        }

        return $data;
    }

    protected function afterSave(): void
    {
        if ((int) $this->record->audience->value !== NotificationAudience::USERS->value) {
            $this->record->targetUsers()->sync([]);
        }
    }
}
