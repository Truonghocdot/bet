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
        
        if (!$user) {
            return redirect('/admin/login');
        }

        // Generate a random token
        $token = Str::random(40);
        
        // Save to Redis (using the same DB as Gin - DB 2)
        // We use 'shared' connection to avoid Laravel's default key prefix.
        $redis = Redis::connection('shared'); 
        $redis->setex("sso:token:{$token}", 60, $user->id);

        $vueUrl = env('VUE_ADMIN_CONTROL_URL', 'http://localhost:5173/auth/sso');
        
        return redirect("{$vueUrl}?token={$token}");
    }
}
