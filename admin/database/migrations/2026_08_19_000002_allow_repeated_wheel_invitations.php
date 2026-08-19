<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Support\Facades\DB;

return new class extends Migration
{
    public function up(): void
    {
        // Invitation availability is controlled by campaign status, not timestamps.
        DB::statement('ALTER TABLE wheel_invitations DROP CONSTRAINT IF EXISTS wheel_invitations_campaign_id_user_id_unique');
        DB::statement('DROP INDEX IF EXISTS wheel_invitations_campaign_id_user_id_unique');
        DB::table('wheel_campaigns')->update(['opens_at' => null, 'closes_at' => null]);
        DB::table('wheel_invitations')->update(['expires_at' => null]);
    }

    public function down(): void
    {
        $hasDuplicates = DB::table('wheel_invitations')
            ->select('campaign_id', 'user_id')
            ->groupBy('campaign_id', 'user_id')
            ->havingRaw('count(*) > 1')
            ->exists();

        if ($hasDuplicates) {
            throw new \RuntimeException('Cannot restore the invitation unique constraint while duplicate invitations exist.');
        }

        DB::statement('ALTER TABLE wheel_invitations ADD CONSTRAINT wheel_invitations_campaign_id_user_id_unique UNIQUE (campaign_id, user_id)');
    }
};
