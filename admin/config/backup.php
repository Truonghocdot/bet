<?php

return [
    'disk' => env('BACKUP_DISK', 'r2'),
    'site' => env('BACKUP_SITE'),
    'prefix' => env('BACKUP_PREFIX', 'backups'),

    'tables' => [
        'users',
        'wallets',
        'payment_receiving_accounts',
        'transactions',
        'wallet_ledger_entries',
        'affiliate_reward_settings',
        'affiliate_profiles',
        'affiliate_links',
        'affiliate_referrals',
        'affiliate_reward_logs',
        'news_articles',
        'banners',
    ],

    'asset_disk' => env('BACKUP_ASSET_DISK', 'public'),
    'asset_directories' => [
        'news',
        'banners',
    ],

    'retention_days' => (int) env('BACKUP_RETENTION_DAYS', 30),
    'keep_minimum' => (int) env('BACKUP_KEEP_MINIMUM', 7),
    'schedule' => env('BACKUP_SCHEDULE', '02:15'),
];
