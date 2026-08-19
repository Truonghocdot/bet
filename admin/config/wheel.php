<?php

return [
    'enabled' => env('WHEEL_EVENT_ENABLED', false),
    'site_code' => env('WHEEL_SITE_CODE', 'fh88u'),
    'microsite_url' => rtrim((string) env('WHEEL_MICROSITE_URL', 'http://localhost:5173/event.html'), '/'),
    'duration_seconds' => 300,
    'spin_duration_seconds' => 5,
];
