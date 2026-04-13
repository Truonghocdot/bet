<?php

use App\Enum\Notification\NotificationAudience;
use App\Enum\Notification\NotificationStatus;
use Illuminate\Database\Migrations\Migration;
use Illuminate\Support\Facades\DB;

return new class extends Migration
{
    public function up(): void
    {
        DB::table('notifications')
            ->whereNotIn('status', array_map(static fn (NotificationStatus $status): int => $status->value, NotificationStatus::cases()))
            ->update(['status' => NotificationStatus::DRAFT->value]);

        DB::table('notifications')
            ->whereNotIn('audience', array_map(static fn (NotificationAudience $audience): int => $audience->value, NotificationAudience::cases()))
            ->update(['audience' => NotificationAudience::ALL->value]);

        DB::statement('ALTER TABLE notifications ALTER COLUMN status SET DEFAULT '.NotificationStatus::DRAFT->value);
        DB::statement('ALTER TABLE notifications ALTER COLUMN audience SET DEFAULT '.NotificationAudience::ALL->value);

        DB::statement('ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_status_check');
        DB::statement(sprintf(
            'ALTER TABLE notifications ADD CONSTRAINT notifications_status_check CHECK (status IN (%s))',
            implode(',', array_map(static fn (NotificationStatus $status): int => $status->value, NotificationStatus::cases()))
        ));

        DB::statement('ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_audience_check');
        DB::statement(sprintf(
            'ALTER TABLE notifications ADD CONSTRAINT notifications_audience_check CHECK (audience IN (%s))',
            implode(',', array_map(static fn (NotificationAudience $audience): int => $audience->value, NotificationAudience::cases()))
        ));

        DB::statement('ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_publish_expires_check');
        DB::statement('UPDATE notifications SET expires_at = NULL WHERE expires_at IS NOT NULL AND publish_at IS NOT NULL AND expires_at <= publish_at');
        DB::statement('ALTER TABLE notifications ADD CONSTRAINT notifications_publish_expires_check CHECK (expires_at IS NULL OR publish_at IS NULL OR expires_at > publish_at)');
    }

    public function down(): void
    {
        DB::statement('ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_status_check');
        DB::statement('ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_audience_check');
        DB::statement('ALTER TABLE notifications DROP CONSTRAINT IF EXISTS notifications_publish_expires_check');
        DB::statement('ALTER TABLE notifications ALTER COLUMN status DROP DEFAULT');
        DB::statement('ALTER TABLE notifications ALTER COLUMN audience DROP DEFAULT');
    }
};
