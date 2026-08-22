<?php

namespace Tests\Feature;

use App\Filament\Resources\Chat\ChatMessages\ChatMessageResource;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Schema;
use Tests\TestCase;

class ChatMessageResourceQueryTest extends TestCase
{
    protected function setUp(): void
    {
        parent::setUp();

        Schema::create('chat_rooms', function (Blueprint $table): void {
            $table->id();
            $table->unsignedBigInteger('wheel_invitation_id')->nullable();
            $table->unsignedBigInteger('wheel_session_id')->nullable();
            $table->string('code');
            $table->string('name');
            $table->boolean('enabled')->default(true);
            $table->timestamps();
        });
        Schema::create('chat_messages', function (Blueprint $table): void {
            $table->id();
            $table->unsignedBigInteger('room_id');
            $table->string('actor_type');
            $table->unsignedBigInteger('user_id')->nullable();
            $table->unsignedBigInteger('bot_profile_id')->nullable();
            $table->string('display_name');
            $table->string('body');
            $table->unsignedTinyInteger('status')->default(1);
            $table->unsignedBigInteger('moderated_by')->nullable();
            $table->timestamp('moderated_at')->nullable();
            $table->timestamps();
        });

        DB::table('chat_rooms')->insert([
            ['id' => 1, 'wheel_invitation_id' => null, 'wheel_session_id' => null, 'code' => 'global', 'name' => 'Global', 'created_at' => now(), 'updated_at' => now()],
            ['id' => 2, 'wheel_invitation_id' => 44, 'wheel_session_id' => null, 'code' => 'wheel-44', 'name' => 'Event', 'created_at' => now(), 'updated_at' => now()],
        ]);
        DB::table('chat_messages')->insert([
            ['id' => 1, 'room_id' => 1, 'actor_type' => 'user', 'user_id' => 10, 'bot_profile_id' => null, 'display_name' => 'ID game #10', 'body' => 'Global user', 'created_at' => now(), 'updated_at' => now()],
            ['id' => 2, 'room_id' => 2, 'actor_type' => 'bot', 'user_id' => null, 'bot_profile_id' => 20, 'display_name' => 'ID game #20', 'body' => 'Event bot', 'created_at' => now(), 'updated_at' => now()],
            ['id' => 3, 'room_id' => 2, 'actor_type' => 'user', 'user_id' => 30, 'bot_profile_id' => null, 'display_name' => 'ID game #30', 'body' => 'Event user', 'created_at' => now(), 'updated_at' => now()],
        ]);
    }

    protected function tearDown(): void
    {
        Schema::dropIfExists('chat_messages');
        Schema::dropIfExists('chat_rooms');

        parent::tearDown();
    }

    public function test_admin_message_query_only_returns_invited_user_messages(): void
    {
        self::assertSame([3], ChatMessageResource::getEloquentQuery()->pluck('id')->all());
    }
}
