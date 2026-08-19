<?php

namespace App\Services\Wheel;

use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Redis;
use Throwable;

class WheelEventPublisher
{
    public function queue(string $topic, string $event, array $payload): int
    {
        return (int) DB::table('wheel_outbox_events')->insertGetId([
            'topic' => $topic,
            'event' => $event,
            'payload' => json_encode($payload, JSON_THROW_ON_ERROR),
            'available_at' => now(),
            'created_at' => now(),
            'updated_at' => now(),
        ]);
    }

    public function queueForUser(int $userId, string $event, array $payload): int
    {
        return $this->queue('stream:user:'.$userId, $event, $payload);
    }

    public function queueForSession(int $sessionId, string $event, array $payload): int
    {
        return $this->queue('stream:wheel:session:'.$sessionId, $event, $payload);
    }

    public function publishPending(int $limit = 100): int
    {
        $published = 0;
        $rows = DB::table('wheel_outbox_events')
            ->whereNull('published_at')
            ->where(fn ($query) => $query->whereNull('available_at')->orWhere('available_at', '<=', now()))
            ->orderBy('id')
            ->limit($limit)
            ->get();

        foreach ($rows as $row) {
            try {
                $payload = json_decode((string) $row->payload, true, flags: JSON_THROW_ON_ERROR);
                Redis::connection('shared')->publish((string) $row->topic, json_encode([
                    'event' => (string) $row->event,
                    'data' => $payload,
                    'published_at' => now()->toISOString(),
                ], JSON_THROW_ON_ERROR));
                DB::table('wheel_outbox_events')->where('id', $row->id)->whereNull('published_at')->update([
                    'published_at' => now(),
                    'attempts' => ((int) $row->attempts) + 1,
                    'last_error' => null,
                    'updated_at' => now(),
                ]);
                $published++;
            } catch (Throwable $exception) {
                $attempts = ((int) $row->attempts) + 1;
                DB::table('wheel_outbox_events')->where('id', $row->id)->update([
                    'attempts' => $attempts,
                    'last_error' => mb_substr($exception->getMessage(), 0, 2000),
                    'available_at' => now()->addSeconds(min(300, 2 ** min($attempts, 8))),
                    'updated_at' => now(),
                ]);
            }
        }

        return $published;
    }
}
