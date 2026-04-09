<?php

namespace Database\Seeders;

use App\Enum\User\RoleUser;
use App\Enum\User\UserStatus;
use App\Models\User;
use Illuminate\Database\Console\Seeds\WithoutModelEvents;
use Illuminate\Database\Seeder;
use Illuminate\Support\Facades\Hash;

class DatabaseSeeder extends Seeder
{
    use WithoutModelEvents;

    public function run(): void
    {
        User::updateOrCreate(
            ['email' => 'admin@ff789.club'],
            [
                'name' => 'Administrator',
                'phone' => null,
                'password' => Hash::make('password'),
                'role' => RoleUser::ADMIN,
                'status' => UserStatus::ACTIVE,
            ],
        );

        $this->call([
            ExchangeRateSettingSeeder::class,
            VietQrBankSeeder::class,
        ]);
    }
}
