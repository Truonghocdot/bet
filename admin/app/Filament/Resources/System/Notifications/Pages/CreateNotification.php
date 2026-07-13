<?php

namespace App\Filament\Resources\System\Notifications\Pages;

use App\Enum\Notification\NotificationAudience;
use App\Enum\Notification\NotificationStatus;
use App\Filament\Resources\System\Notifications\NotificationResource;
use App\Filament\Resources\System\Notifications\Support\NotificationTargetSyncer;
use Filament\Resources\Pages\CreateRecord;
use Illuminate\Validation\ValidationException;

class CreateNotification extends CreateRecord
{
    protected static string $resource = NotificationResource::class;

    /**
     * @var list<int|string>
     */
    private array $targetUserIds = [];

    protected function mutateFormDataBeforeCreate(array $data): array
    {
        $data = $this->normalizeAndValidate($data);
        $data['created_by'] = auth()->id();

        return $data;
    }

    protected function afterCreate(): void
    {
        NotificationTargetSyncer::sync($this->record, $this->targetUserIds);
    }

    private function normalizeAndValidate(array $data): array
    {
        $status = (int) ($data['status'] ?? 0);
        $audience = (int) ($data['audience'] ?? 0);
        $this->targetUserIds = array_values($data['targetUsers'] ?? []);
        unset($data['targetUsers']);

        $data['body'] = $this->normalizeBody($data['body'] ?? null);

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

        if (blank($data['image_path'] ?? null) && blank($data['body'])) {
            throw ValidationException::withMessages([
                'body' => 'Vui lòng nhập nội dung hoặc tải lên ảnh thông báo.',
            ]);
        }

        if ($audience === NotificationAudience::USERS->value && empty($this->targetUserIds)) {
            throw ValidationException::withMessages([
                'targetUsers' => 'Vui lòng chọn ít nhất 1 người dùng đích.',
            ]);
        }

        return $data;
    }

    private function normalizeBody(mixed $value): ?string
    {
        $normalized = trim((string) ($value ?? ''));

        return $normalized === '' ? null : $normalized;
    }
}
