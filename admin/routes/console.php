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

Schedule::command('deposits:expire-pending')
    ->everyMinute()
    ->withoutOverlapping();
