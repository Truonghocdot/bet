<?php

namespace App\Providers;

use Filament\Support\Facades\FilamentTimezone;
use Illuminate\Support\Facades\URL;
use Illuminate\Support\ServiceProvider;

class AppServiceProvider extends ServiceProvider
{
    /**
     * Register any application services.
     */
    public function register(): void
    {
        //
    }

    /**
     * Bootstrap any application services.
     */
    public function boot(): void
    {
        FilamentTimezone::set(config('app.timezone', 'Asia/Ho_Chi_Minh'));

        if (config('app.env') === 'local' && (isset($_SERVER['HTTP_X_FORWARDED_PROTO']) || isset($_SERVER['HTTP_X_FORWARDED_HOST']))) {
            URL::forceScheme('https');
        }
    }
}
