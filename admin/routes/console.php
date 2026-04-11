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

Schedule::command('banks:sync-vietqr')
    ->dailyAt('03:15')
    ->withoutOverlapping();

Schedule::command('payment:prime-receiving-accounts')
    ->everyFiveMinutes()
    ->withoutOverlapping();

Schedule::command('bet:cleanup')
    ->dailyAt('03:45')
    ->withoutOverlapping();
