<?php

return [
    'enabled' => env('LOG_VIEWER_ENABLED', true),

    'api_only' => env('LOG_VIEWER_API_ONLY', false),

    // false = public in production, true = package sẽ yêu cầu auth/gate
    'require_auth_in_production' => env('LOG_VIEWER_REQUIRE_AUTH_IN_PRODUCTION', false),

    'route_domain' => null,

    'route_path' => env('LOG_VIEWER_PATH', 'log-viewer'),

    'assets_path' => 'vendor/log-viewer',

    'back_to_system_url' => config('app.url', null),

    'back_to_system_label' => null,

    'timezone' => null,

    'datetime_format' => 'Y-m-d H:i:s',
];
