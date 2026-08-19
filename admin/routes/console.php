<?php

use Illuminate\Foundation\Inspiring;
use Illuminate\Support\Facades\Artisan;
use Illuminate\Support\Facades\Schedule;

Artisan::command('inspire', function () {
    $this->comment(Inspiring::quote());
})->purpose('Display an inspiring quote');

Schedule::command('rates:sync-usdt-vnd')
    ->everyFiveMinutes()
    ->withoutOverlapping();

Schedule::command('payment:prime-receiving-accounts')
    ->everyFiveMinutes()
    ->withoutOverlapping();

Schedule::command('finance:generate-fake-transactions')
    ->everyFiveSeconds()
    ->withoutOverlapping();

Schedule::command('chat:generate-bot-message')
    ->everyTwentySeconds()
    ->withoutOverlapping();

Schedule::command('chat:prune')
    ->everyMinute()
    ->withoutOverlapping();

Schedule::command('wheel:maintain')
    ->everyMinute()
    ->withoutOverlapping();

Schedule::command('wheel:publish-outbox')
    ->everyFiveSeconds()
    ->withoutOverlapping();

Schedule::command('backup:create --prune')
    ->dailyAt((string) config('backup.schedule', '02:15'))
    ->withoutOverlapping(180);
