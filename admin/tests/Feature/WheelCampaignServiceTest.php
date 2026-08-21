<?php

namespace Tests\Feature;

use App\Models\Wheel\WheelCampaign;
use App\Models\Wheel\WheelInvitation;
use App\Services\Wheel\WheelCampaignService;
use App\Services\Wheel\WheelEventPublisher;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Schema;
use Illuminate\Validation\ValidationException;
use Mockery;
use Tests\TestCase;

class WheelCampaignServiceTest extends TestCase
{
    protected function setUp(): void
    {
        parent::setUp();

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
        Schema::dropIfExists('wheel_invitation_rounds');
        Schema::dropIfExists('wheel_invitations');
        Schema::dropIfExists('wheel_campaign_round_templates');
        Schema::dropIfExists('wheel_campaigns');
        Schema::dropIfExists('users');
        parent::tearDown();
    }

    public function test_it_snapshots_four_rounds_and_activates_selected_user(): void
    {
        $campaign = $this->campaign();
        DB::table('users')->insert(['id' => 203985, 'name' => 'Khách thử', 'phone' => '+84900000000', 'role' => 2]);

        $count = $this->service()->inviteUsers($campaign, [203985]);

        self::assertSame(1, $count);
        $invitation = WheelInvitation::query()->with('rounds')->firstOrFail();
        self::assertSame('pending', $invitation->status);
        self::assertCount(4, $invitation->rounds);
        self::assertSame('50000000', rtrim(rtrim((string) $invitation->rounds->firstWhere('round_no', 2)?->prize_amount, '0'), '.'));
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
            [2, 'jackpot_50m', 'Giải 50 triệu', '50000000'],
            [3, 'try_again', 'May mắn lần sau', '0'],
            [4, 'thank_you', 'Cảm ơn', '0'],
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
