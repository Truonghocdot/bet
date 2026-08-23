<?php

namespace Tests\Feature;

use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Schema;
use Tests\TestCase;

class WheelThreeRoundMigrationTest extends TestCase
{
    protected function setUp(): void
    {
        parent::setUp();

        Schema::create('wheel_campaign_round_templates', function (Blueprint $table): void {
            $table->id();
            $table->unsignedBigInteger('campaign_id');
            $table->unsignedTinyInteger('round_no');
            $table->string('segment_key');
            $table->string('result_label');
            $table->decimal('prize_amount', 30, 8);
            $table->timestamps();
        });
        Schema::create('wheel_invitations', function (Blueprint $table): void {
            $table->id();
            $table->string('status');
            $table->timestamps();
        });
        Schema::create('wheel_invitation_rounds', function (Blueprint $table): void {
            $table->id();
            $table->unsignedBigInteger('invitation_id');
            $table->unsignedTinyInteger('round_no');
            $table->string('segment_key');
            $table->string('result_label');
            $table->decimal('prize_amount', 30, 8);
            $table->string('status');
            $table->timestamps();
        });
        Schema::create('wheel_sessions', function (Blueprint $table): void {
            $table->id();
            $table->unsignedBigInteger('invitation_id');
            $table->string('status');
            $table->unsignedTinyInteger('current_round');
            $table->timestamp('completed_at')->nullable();
            $table->unsignedInteger('version')->default(1);
            $table->timestamps();
        });
        Schema::create('chat_rooms', function (Blueprint $table): void {
            $table->id();
            $table->unsignedBigInteger('wheel_session_id')->nullable();
            $table->boolean('enabled')->default(true);
            $table->timestamps();
        });
    }

    protected function tearDown(): void
    {
        Schema::dropIfExists('chat_rooms');
        Schema::dropIfExists('wheel_sessions');
        Schema::dropIfExists('wheel_invitation_rounds');
        Schema::dropIfExists('wheel_invitations');
        Schema::dropIfExists('wheel_campaign_round_templates');

        parent::tearDown();
    }

    public function test_it_converts_pending_round_four_sessions_to_the_three_round_format(): void
    {
        foreach (range(1, 4) as $round) {
            DB::table('wheel_campaign_round_templates')->insert([
                'campaign_id' => 1,
                'round_no' => $round,
                'segment_key' => $round === 2 ? 'jackpot_50m' : 'try_again',
                'result_label' => $round === 2 ? '50 triệu' : 'May mắn',
                'prize_amount' => $round === 2 ? 50000000 : 0,
                'created_at' => now(),
                'updated_at' => now(),
            ]);
            DB::table('wheel_invitation_rounds')->insert([
                'invitation_id' => 7,
                'round_no' => $round,
                'segment_key' => $round === 2 ? 'jackpot_50m' : 'try_again',
                'result_label' => $round === 2 ? '50 triệu' : 'May mắn',
                'prize_amount' => $round === 2 ? 50000000 : 0,
                'status' => $round < 4 ? 'spun' : 'pending',
                'created_at' => now(),
                'updated_at' => now(),
            ]);
            DB::table('wheel_invitation_rounds')->insert([
                'invitation_id' => 8,
                'round_no' => $round,
                'segment_key' => $round === 2 ? 'jackpot_50m' : 'try_again',
                'result_label' => $round === 2 ? '50 triệu' : 'May mắn',
                'prize_amount' => $round === 2 ? 50000000 : 0,
                'status' => 'pending',
                'created_at' => now(),
                'updated_at' => now(),
            ]);
        }
        DB::table('wheel_invitations')->insert(['id' => 7, 'status' => 'started', 'created_at' => now(), 'updated_at' => now()]);
        DB::table('wheel_invitations')->insert(['id' => 8, 'status' => 'pending', 'created_at' => now(), 'updated_at' => now()]);
        DB::table('wheel_sessions')->insert(['id' => 9, 'invitation_id' => 7, 'status' => 'active', 'current_round' => 4, 'version' => 1, 'created_at' => now(), 'updated_at' => now()]);
        DB::table('chat_rooms')->insert(['wheel_session_id' => 9, 'enabled' => true, 'created_at' => now(), 'updated_at' => now()]);

        $migration = require database_path('migrations/2026_08_23_000001_convert_wheel_to_three_rounds.php');
        $migration->up();

        self::assertSame(3, DB::table('wheel_campaign_round_templates')->where('campaign_id', 1)->count());
        self::assertDatabaseHas('wheel_campaign_round_templates', ['campaign_id' => 1, 'round_no' => 2, 'segment_key' => 'reward_39m', 'prize_amount' => 39000000]);
        self::assertSame(3, DB::table('wheel_invitation_rounds')->where('invitation_id', 7)->count());
        self::assertSame(3, DB::table('wheel_invitation_rounds')->where('invitation_id', 8)->count());
        self::assertDatabaseHas('wheel_invitation_rounds', ['invitation_id' => 8, 'round_no' => 2, 'segment_key' => 'reward_39m', 'prize_amount' => 39000000]);
        self::assertDatabaseHas('wheel_sessions', ['id' => 9, 'status' => 'completed', 'current_round' => 3]);
        self::assertDatabaseHas('wheel_invitations', ['id' => 7, 'status' => 'completed']);
        self::assertDatabaseHas('chat_rooms', ['wheel_session_id' => 9, 'enabled' => false]);
    }
}
