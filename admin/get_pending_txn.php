<?php
require __DIR__.'/vendor/autoload.php';
$app = require_once __DIR__.'/bootstrap/app.php';
$kernel = $app->make(Illuminate\Contracts\Console\Kernel::class);
$kernel->bootstrap();

use Illuminate\Support\Facades\DB;

$txn = DB::table('transactions')
    ->where('status', 1)
    ->whereNotNull('client_ref')
    ->orderBy('id', 'desc')
    ->first();

echo json_encode($txn) . "\n";
