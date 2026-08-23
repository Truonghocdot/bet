<?php

namespace Tests\Feature;

use App\Jobs\GenerateChatBotMessage;
use App\Models\Chat\ChatRoom;
use App\Services\Chat\ChatRedisPublisher;
use Carbon\CarbonImmutable;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Queue;
use Illuminate\Support\Facades\Schema;
use Mockery;
use Tests\TestCase;

class GenerateWheelBotMessageTest extends TestCase
{
    protected function setUp(): void
    {
        parent::setUp();

        config(['wheel.enabled' => true, 'cache.default' => 'array']);
        Cache::setDefaultDriver('array');
        Queue::fake();

        Schema::create('wheel_invitations', function (Blueprint $table): void {
            $table->id();
            $table->string('status');
            $table->boolean('bot_chat_enabled')->default(true);
            $table->timestamps();
        });
        Schema::create('wheel_sessions', function (Blueprint $table): void {
            $table->id();
            $table->unsignedBigInteger('invitation_id');
            $table->string('status');
            $table->timestamp('ends_at');
            $table->timestamps();
        });
        Schema::create('chat_rooms', function (Blueprint $table): void {
            $table->id();
            $table->unsignedBigInteger('wheel_invitation_id')->nullable();
            $table->unsignedBigInteger('wheel_session_id')->nullable();
            $table->string('code')->unique();
            $table->string('name');
            $table->boolean('enabled')->default(true);
            $table->timestamp('next_bot_at')->nullable();
            $table->unsignedSmallInteger('bot_message_count')->default(0);
            $table->timestamps();
        });
        Schema::create('chat_bot_profiles', function (Blueprint $table): void {
            $table->id();
            $table->string('display_name');
            $table->string('avatar_path')->nullable();
            $table->boolean('active')->default(true);
            $table->unsignedInteger('sort_order')->default(0);
            $table->timestamps();
        });
        Schema::create('chat_bot_templates', function (Blueprint $table): void {
            $table->id();
            $table->unsignedBigInteger('bot_profile_id')->nullable();
            $table->string('body', 280);
            $table->string('category')->default('general');
            $table->string('language')->default('vi');
            $table->boolean('active')->default(true);
            $table->timestamp('last_used_at')->nullable();
            $table->unsignedInteger('usage_count')->default(0);
            $table->timestamps();
        });
        Schema::create('chat_messages', function (Blueprint $table): void {
            $table->id();
            $table->unsignedBigInteger('room_id');
            $table->string('actor_type');
            $table->unsignedBigInteger('user_id')->nullable();
            $table->unsignedBigInteger('bot_profile_id')->nullable();
            $table->string('display_name');
            $table->string('body', 280);
            $table->unsignedTinyInteger('status')->default(1);
            $table->unsignedBigInteger('moderated_by')->nullable();
            $table->timestamp('moderated_at')->nullable();
            $table->timestamps();
        });
    }

    protected function tearDown(): void
    {
        Schema::dropIfExists('chat_messages');
        Schema::dropIfExists('chat_bot_templates');
        Schema::dropIfExists('chat_bot_profiles');
        Schema::dropIfExists('chat_rooms');
        Schema::dropIfExists('wheel_sessions');
        Schema::dropIfExists('wheel_invitations');

        parent::tearDown();
    }

    public function test_it_seeds_four_utc_messages_before_the_session_starts(): void
    {
        $room = $this->createPendingRoom();

        $publisher = Mockery::mock(ChatRedisPublisher::class);
        $publisher->shouldReceive('publishWheelInvitation')->times(4);

        (new GenerateChatBotMessage($room->id))->handle($publisher);

        self::assertSame(4, DB::table('chat_messages')->where('room_id', $room->id)->where('actor_type', 'bot')->count());
        self::assertSame(4, (int) $room->fresh()->bot_message_count);
        $createdAt = CarbonImmutable::parse((string) DB::table('chat_messages')->min('created_at'), 'UTC');
        self::assertLessThanOrEqual(5, abs($createdAt->diffInSeconds(now('UTC'), false)));
        $nextBotAt = CarbonImmutable::parse((string) $room->fresh()->getRawOriginal('next_bot_at'), 'UTC');
        self::assertLessThanOrEqual(2, $nextBotAt->diffInSeconds(now('UTC')));
        Queue::assertPushed(GenerateChatBotMessage::class, fn (GenerateChatBotMessage $job): bool => $job->wheelRoomId === $room->id);
    }

    public function test_it_completes_the_opening_burst_when_one_message_already_exists(): void
    {
        $room = $this->createPendingRoom(1);
        DB::table('chat_messages')->insert([
            'room_id' => $room->id,
            'actor_type' => 'bot',
            'bot_profile_id' => 1,
            'display_name' => 'Người chơi 1',
            'body' => 'Tin nhắn đã có',
            'status' => 1,
            'created_at' => now('UTC'),
            'updated_at' => now('UTC'),
        ]);

        $publisher = Mockery::mock(ChatRedisPublisher::class);
        $publisher->shouldReceive('publishWheelInvitation')->times(3);

        (new GenerateChatBotMessage($room->id))->handle($publisher);

        self::assertSame(4, DB::table('chat_messages')->where('room_id', $room->id)->where('actor_type', 'bot')->count());
        self::assertSame(4, (int) $room->fresh()->bot_message_count);
    }

    private function createPendingRoom(int $botMessageCount = 0): ChatRoom
    {
        DB::table('wheel_invitations')->insert([
            'id' => 11,
            'status' => 'pending',
            'bot_chat_enabled' => true,
            'created_at' => now(),
            'updated_at' => now(),
        ]);
        foreach (range(1, 4) as $index) {
            DB::table('chat_bot_profiles')->insert([
                'id' => $index,
                'display_name' => 'Người chơi '.$index,
                'active' => true,
                'sort_order' => $index,
                'created_at' => now(),
                'updated_at' => now(),
            ]);
            DB::table('chat_bot_templates')->insert([
                'body' => 'Tin nhắn mở đầu '.$index,
                'active' => true,
                'created_at' => now(),
                'updated_at' => now(),
            ]);
        }

        return ChatRoom::query()->create([
            'wheel_invitation_id' => 11,
            'code' => 'wheel-invitation-11',
            'name' => 'Phòng thử nghiệm',
            'enabled' => true,
            'next_bot_at' => now('UTC'),
            'bot_message_count' => $botMessageCount,
        ]);
    }
}
