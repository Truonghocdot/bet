<?php

return [

    'postmark' => [
        'key' => env('POSTMARK_API_KEY'),
    ],

    'resend' => [
        'key' => env('RESEND_API_KEY'),
    ],

    'ses' => [
        'key' => env('AWS_ACCESS_KEY_ID'),
        'secret' => env('AWS_SECRET_ACCESS_KEY'),
        'region' => env('AWS_DEFAULT_REGION', 'us-east-1'),
    ],

    'slack' => [
        'notifications' => [
            'bot_user_oauth_token' => env('SLACK_BOT_USER_OAUTH_TOKEN'),
            'channel' => env('SLACK_BOT_USER_DEFAULT_CHANNEL'),
        ],
    ],

    'exchange_rate' => [
        'usdt_vnd_url' => env('USDT_VND_RATE_URL', 'https://api.coingecko.com/api/v3/simple/price?ids=tether&vs_currencies=vnd'),
        'source_name' => env('USDT_VND_RATE_SOURCE_NAME', 'coingecko'),
        'timeout' => (int) env('USDT_VND_RATE_TIMEOUT', 10),
        'retry' => (int) env('USDT_VND_RATE_RETRY', 2),
        'cache_ttl_seconds' => (int) env('USDT_VND_RATE_CACHE_TTL_SECONDS', 300),
        'cache_store' => env('USDT_VND_RATE_CACHE_STORE', 'redis'),
        'cache_key' => env('USDT_VND_RATE_CACHE_KEY', 'admin:exchange-rate:usdt-vnd:snapshot'),
        'redis_connection' => env('USDT_VND_RATE_REDIS_CONNECTION', 'shared'),
        'redis_key' => env('USDT_VND_RATE_REDIS_KEY', 'shared:exchange-rate:usdt-vnd'),
    ],

    'vietqr_banks' => [
        'url' => env('VIETQR_BANKS_URL', 'https://api.vietqr.io/v2/banks'),
        'source_name' => env('VIETQR_BANKS_SOURCE_NAME', 'vietqr'),
        'timeout' => (int) env('VIETQR_BANKS_TIMEOUT', 10),
        'retry' => (int) env('VIETQR_BANKS_RETRY', 2),
        'cache_ttl_seconds' => (int) env('VIETQR_BANKS_CACHE_TTL_SECONDS', 86400),
        'cache_store' => env('VIETQR_BANKS_CACHE_STORE', 'redis'),
        'cache_key' => env('VIETQR_BANKS_CACHE_KEY', 'admin:vietqr:banks:snapshot'),
        'redis_connection' => env('VIETQR_BANKS_REDIS_CONNECTION', 'shared'),
        'redis_key' => env('VIETQR_BANKS_REDIS_KEY', 'shared:vietqr:banks'),
    ],

];
