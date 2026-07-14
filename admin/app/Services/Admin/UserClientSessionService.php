<?php

namespace App\Services\Admin;

use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Redis;
use Throwable;

class UserClientSessionService
{
    private const REDIS_CONNECTION = 'shared';

    /**
     * @var list<string>
     */
    private const CLIENT_SCOPES = ['main', 'agency'];

    public function invalidateSessions(int $userId): void
    {
        DB::table('auth_refresh_tokens')
            ->where('user_id', $userId)
            ->delete();

        try {
            $redis = Redis::connection(self::REDIS_CONNECTION);

            foreach (self::CLIENT_SCOPES as $scope) {
                $redis->del(sprintf('user:session:%s:%d', $scope, $userId));
            }
        } catch (Throwable $exception) {
            Log::warning('Failed to invalidate client sessions for user.', [
                'user_id' => $userId,
                'exception' => $exception->getMessage(),
            ]);
        }
    }
}
