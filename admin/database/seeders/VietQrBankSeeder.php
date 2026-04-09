<?php

namespace Database\Seeders;

use App\Services\Admin\VietQrBankService;
use Illuminate\Database\Seeder;

class VietQrBankSeeder extends Seeder
{
    public function run(VietQrBankService $service): void
    {
        $service->syncFromProvider();
    }
}
