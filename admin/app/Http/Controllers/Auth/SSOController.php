<?php

namespace App\Http\Controllers\Auth;

use App\Http\Controllers\Controller;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Redis;
use Illuminate\Support\Str;

class SSOController extends Controller
{
    public function redirect(Request $request)
    {
        $user = auth()->user();

        if (! $user) {
            return redirect('/admin/login');
        }

        // Generate a random token
        $token = Str::random(40);

        // Save to Redis (using the same DB as Gin - DB 2)
        // We use 'shared' connection to avoid Laravel's default key prefix.
        $redis = Redis::connection('shared');
        $redis->setex("sso:token:{$token}", 60, $user->id);

        $vueUrl = $this->resolveControlUrl($request);

        $query = [
            'token' => $token,
        ];

        $roomCode = trim((string) $request->query('room_code', ''));
        if ($roomCode !== '') {
            $query['room_code'] = $roomCode;
        }

        return redirect($vueUrl.'?'.http_build_query($query));
    }

    private function resolveControlUrl(Request $request): string
    {
        $configured = trim((string) env('VUE_ADMIN_CONTROL_URL', ''));
        if ($configured !== '') {
            $path = trim((string) parse_url($configured, PHP_URL_PATH));

            return $path === ''
                ? rtrim($configured, '/').'/auth/sso'
                : $configured;
        }

        $scheme = $request->getScheme();
        $host = $request->getHost();
        $port = $request->getPort();
        $portSuffix = in_array($port, [80, 443], true) ? '' : ':'.$port;

        if (str_starts_with($host, 'admin.')) {
            return sprintf('%s://%s%s/auth/sso', $scheme, substr($host, 6), $portSuffix);
        }

        return sprintf('%s://%s%s/auth/sso', $scheme, $host, $portSuffix);
    }
}
