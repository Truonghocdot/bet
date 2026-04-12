<?php

use App\Http\Controllers\Auth\SSOController;
use Illuminate\Support\Facades\Route;

Route::get('/auth/sso/redirect', [SSOController::class, 'redirect'])->middleware(['auth'])->name('auth.sso.redirect');
