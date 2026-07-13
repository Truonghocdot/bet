<?php

use App\Enum\Notification\NotificationResponseStatus;
use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('notifications', function (Blueprint $table): void {
            $table->string('image_path')->nullable()->after('body');
        });

        DB::statement('ALTER TABLE notifications ALTER COLUMN body DROP NOT NULL');

        Schema::table('notification_targets', function (Blueprint $table): void {
            $table->unsignedTinyInteger('response_status')->nullable()->after('user_id');
            $table->timestamp('responded_at')->nullable()->after('response_status');
            $table->index(['notification_id', 'response_status'], 'notification_targets_notification_response_status_idx');
        });

        DB::statement('ALTER TABLE notification_targets DROP CONSTRAINT IF EXISTS notification_targets_response_status_check');
        DB::statement(sprintf(
            'ALTER TABLE notification_targets ADD CONSTRAINT notification_targets_response_status_check CHECK (response_status IS NULL OR response_status IN (%s))',
            implode(',', array_map(static fn (NotificationResponseStatus $status): int => $status->value, NotificationResponseStatus::cases()))
        ));
    }

    public function down(): void
    {
        DB::statement('ALTER TABLE notification_targets DROP CONSTRAINT IF EXISTS notification_targets_response_status_check');

        Schema::table('notification_targets', function (Blueprint $table): void {
            $table->dropIndex('notification_targets_notification_response_status_idx');
            $table->dropColumn(['response_status', 'responded_at']);
        });

        DB::statement("UPDATE notifications SET body = '' WHERE body IS NULL");
        DB::statement('ALTER TABLE notifications ALTER COLUMN body SET NOT NULL');

        Schema::table('notifications', function (Blueprint $table): void {
            $table->dropColumn('image_path');
        });
    }
};
