<?php

namespace Tests\Feature;

use App\Models\Wheel\WheelCampaign;
use App\Models\Wheel\WheelInvitation;
use App\Services\Wheel\WheelCampaignService;
use App\Services\Wheel\WheelEventPublisher;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Queue;
use Illuminate\Support\Facades\Schema;
use Illuminate\Validation\ValidationException;
use Mockery;
use Tests\TestCase;

class WheelCampaignServiceTest extends TestCase
{
    protected function setUp(): void
    {
        parent::setUp();
        Queue::fake();

        Schema::create('users', function (Blueprint $table): void {
            $table->unsignedBigInteger('id')->primary();
            $table->string('name');
            $table->string('phone')->nullable();
            $table->unsignedTinyInteger('role');
            $table->softDeletes();
        });
        Schema::create('wheel_campaigns', function (Blueprint $table): void {
            $table->id();
            $table->string('name');
            $table->string('status');
            $table->timestamp('opens_at')->nullable();
            $table->timestamp('closes_at')->nullable();
            $table->unsignedSmallInteger('duration_seconds')->default(300);
            $table->unsignedSmallInteger('spin_duration_seconds')->default(5);
            $table->unsignedBigInteger('created_by')->nullable();
            $table->timestamps();
        });
        Schema::create('wheel_campaign_round_templates', function (Blueprint $table): void {
            $table->id();
            $table->unsignedBigInteger('campaign_id');
            $table->unsignedTinyInteger('round_no');
            $table->string('segment_key');
            $table->string('result_label');
            $table->string('prize_amount');
            $table->timestamps();
        });
        Schema::create('wheel_invitations', function (Blueprint $table): void {
            $table->id();
            $table->string('public_id');
            $table->unsignedBigInteger('campaign_id');
            $table->unsignedBigInteger('user_id');
            $table->string('status');
            $table->boolean('bot_chat_enabled')->default(false);
            $table->timestamp('activated_at')->nullable();
            $table->timestamp('expires_at')->nullable();
            $table->timestamp('popup_seen_at')->nullable();
            $table->timestamp('revoked_at')->nullable();
            $table->unsignedBigInteger('activated_by')->nullable();
            $table->unsignedBigInteger('revoked_by')->nullable();
            $table->timestamps();
        });
        Schema::create('wheel_invitation_rounds', function (Blueprint $table): void {
            $table->id();
            $table->unsignedBigInteger('invitation_id');
            $table->unsignedTinyInteger('round_no');
            $table->string('segment_key');
            $table->string('result_label');
            $table->string('prize_amount');
            $table->string('status');
            $table->timestamp('spun_at')->nullable();
            $table->timestamps();
        });
        Schema::create('wheel_sessions', function (Blueprint $table): void {
            $table->id();
            $table->unsignedBigInteger('invitation_id')->unique();
            $table->unsignedBigInteger('user_id');
            $table->string('status')->default('active');
            $table->timestamps();
        });
        Schema::create('chat_rooms', function (Blueprint $table): void {
            $table->id();
            $table->unsignedBigInteger('wheel_invitation_id')->nullable()->unique();
            $table->unsignedBigInteger('wheel_session_id')->nullable()->unique();
            $table->string('code')->unique();
            $table->string('name');
            $table->boolean('enabled')->default(false);
            $table->timestamp('next_bot_at')->nullable();
            $table->unsignedSmallInteger('bot_message_count')->default(0);
            $table->timestamps();
        });
        Schema::create('wheel_audit_logs', function (Blueprint $table): void {
            $table->id();
            $table->unsignedBigInteger('campaign_id')->nullable();
            $table->unsignedBigInteger('invitation_id')->nullable();
            $table->unsignedBigInteger('session_id')->nullable();
            $table->unsignedBigInteger('actor_user_id')->nullable();
            $table->string('action');
            $table->json('old_values')->nullable();
            $table->json('new_values')->nullable();
            $table->string('ip_address')->nullable();
            $table->timestamp('created_at');
        });
    }

    protected function tearDown(): void
    {
        Schema::dropIfExists('wheel_audit_logs');
        Schema::dropIfExists('chat_rooms');
        Schema::dropIfExists('wheel_sessions');
        Schema::dropIfExists('wheel_invitation_rounds');
        Schema::dropIfExists('wheel_invitations');
        Schema::dropIfExists('wheel_campaign_round_templates');
        Schema::dropIfExists('wheel_campaigns');
        Schema::dropIfExists('users');
        parent::tearDown();
    }

    public function test_it_snapshots_three_rounds_and_activates_selected_user(): void
    {
        $campaign = $this->campaign();
        DB::table('users')->insert(['id' => 203985, 'name' => 'Khách thử', 'phone' => '+84900000000', 'role' => 2]);

        $count = $this->service()->inviteUsers($campaign, [203985]);

        self::assertSame(1, $count);
        $invitation = WheelInvitation::query()->with('rounds')->firstOrFail();
        self::assertSame('pending', $invitation->status);
        self::assertTrue($invitation->bot_chat_enabled);
        self::assertCount(3, $invitation->rounds);
        self::assertDatabaseHas('chat_rooms', ['wheel_invitation_id' => $invitation->id, 'enabled' => true]);
        self::assertSame('39000000', rtrim(rtrim((string) $invitation->rounds->firstWhere('round_no', 2)?->prize_amount, '0'), '.'));
    }

    public function test_snapshot_result_is_locked_after_activation(): void
    {
        $campaign = $this->campaign();
        DB::table('users')->insert(['id' => 203986, 'name' => 'Khách thử 2', 'phone' => '+84900000001', 'role' => 2]);
        $this->service()->inviteUsers($campaign, [203986]);
        $round = WheelInvitation::query()->firstOrFail()->rounds()->firstOrFail();

        $this->expectException(ValidationException::class);
        $round->update(['result_label' => 'Không được phép đổi']);
    }

    public function test_second_round_must_be_the_39m_reward(): void
    {
        $campaign = $this->campaign();
        DB::table('wheel_campaign_round_templates')->where('campaign_id', $campaign->id)->where('round_no', 2)->update([
            'segment_key' => 'reward_68m',
            'prize_amount' => '68000000',
        ]);
        $campaign->load('roundTemplates');
        DB::table('users')->insert(['id' => 203989, 'name' => 'Khách thử 5', 'phone' => '+84900000004', 'role' => 2]);

        $this->expectException(ValidationException::class);
        $this->service()->inviteUsers($campaign, [203989]);
    }

    public function test_second_round_model_normalizes_admin_input_to_39m(): void
    {
        $campaign = $this->campaign();
        $secondRound = $campaign->roundTemplates->firstWhere('round_no', 2);

        $secondRound->update(['segment_key' => 'reward_68m', 'result_label' => '68 triệu', 'prize_amount' => 68000000]);

        self::assertSame('reward_39m', $secondRound->fresh()->segment_key);
        self::assertSame('39000000', rtrim(rtrim((string) $secondRound->fresh()->prize_amount, '0'), '.'));
    }

    public function test_the_same_user_can_receive_multiple_invitations_in_one_campaign(): void
    {
        $campaign = $this->campaign();
        DB::table('users')->insert(['id' => 203987, 'name' => 'Khách thử 3', 'phone' => '+84900000002', 'role' => 2]);

        $count = $this->service()->inviteUsers($campaign, [203987, 203987]);

        self::assertSame(1, $count);
        self::assertCount(1, WheelInvitation::query()->where('campaign_id', $campaign->id)->where('user_id', 203987)->get());

        $this->service()->inviteUsers($campaign, [203987]);

        $invitations = WheelInvitation::query()->where('campaign_id', $campaign->id)->where('user_id', 203987)->get();
        self::assertCount(2, $invitations);
        self::assertNotSame($invitations[0]->public_id, $invitations[1]->public_id);
    }

    public function test_bot_can_be_enabled_for_an_existing_pending_invitation(): void
    {
        $campaign = $this->campaign();
        DB::table('users')->insert(['id' => 203988, 'name' => 'Khách thử 4', 'phone' => '+84900000003', 'role' => 2]);
        $this->service()->inviteUsers($campaign, [203988], true, false);
        $invitation = WheelInvitation::query()->firstOrFail();

        self::assertFalse($invitation->bot_chat_enabled);
        self::assertNull(DB::table('chat_rooms')->where('wheel_invitation_id', $invitation->id)->value('next_bot_at'));

        $this->service()->setBotChatEnabled($invitation, true);

        self::assertTrue($invitation->fresh()->bot_chat_enabled);
        self::assertNotNull(DB::table('chat_rooms')->where('wheel_invitation_id', $invitation->id)->value('next_bot_at'));
    }

    private function campaign(): WheelCampaign
    {
        $campaign = WheelCampaign::query()->create([
            'name' => 'Sự kiện thử nghiệm',
            'status' => 'active',
            'closes_at' => now()->addDay(),
            'duration_seconds' => 300,
            'spin_duration_seconds' => 5,
        ]);
        foreach ([
            [1, 'try_again', 'May mắn lần sau', '0'],
            [2, 'reward_39m', 'Giải 39 triệu', '39000000'],
            [3, 'try_again', 'May mắn lần sau', '0'],
        ] as [$roundNo, $segment, $label, $amount]) {
            $campaign->roundTemplates()->create(['round_no' => $roundNo, 'segment_key' => $segment, 'result_label' => $label, 'prize_amount' => $amount]);
        }

        return $campaign->fresh('roundTemplates');
    }

    private function service(): WheelCampaignService
    {
        $publisher = Mockery::mock(WheelEventPublisher::class);
        $publisher->shouldReceive('queueForUser')->andReturn(1);
        $publisher->shouldReceive('publishPending')->andReturn(1);

        return new WheelCampaignService($publisher);
    }
}
