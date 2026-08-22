<?php

namespace App\Services\Chat;

use App\Models\Chat\ChatMessage;
use Illuminate\Support\Facades\Redis;

class ChatRedisPublisher
{
    public function publish(string $roomCode, string $event, ChatMessage $message): void
    {
        Redis::connection('shared')->publish(
            'stream:chat:global:'.$roomCode,
            json_encode([
                'event' => $event,
                'data' => [
                    'id' => (int) $message->id,
                    'display_name' => $message->display_name,
                    'body' => $message->body,
                    'actor_type' => $message->actor_type,
                    'created_at' => $message->created_at?->toISOString(),
                ],
                'published_at' => now()->toISOString(),
            ], JSON_THROW_ON_ERROR),
        );
    }

    public function publishWheelSession(int $sessionId, string $event, ChatMessage $message): void
    {
        Redis::connection('shared')->publish(
            'stream:wheel:session:'.$sessionId,
            json_encode([
                'event' => $event,
                'data' => [
                    'id' => (int) $message->id,
                    'display_name' => $message->display_name,
                    'body' => $message->body,
                    'actor_type' => $message->actor_type,
                    'created_at' => $message->created_at?->toISOString(),
                ],
                'published_at' => now()->toISOString(),
            ], JSON_THROW_ON_ERROR),
        );
    }

    public function publishWheelInvitation(int $invitationId, string $event, ChatMessage $message): void
    {
        Redis::connection('shared')->publish(
            'stream:wheel:invitation:'.$invitationId,
            json_encode([
                'event' => $event,
                'data' => [
                    'id' => (int) $message->id,
                    'display_name' => $message->display_name,
                    'body' => $message->body,
                    'actor_type' => $message->actor_type,
                    'created_at' => $message->created_at?->toISOString(),
                ],
                'published_at' => now()->toISOString(),
            ], JSON_THROW_ON_ERROR),
        );
    }
}
