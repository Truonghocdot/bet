<?php

namespace Database\Seeders;

use App\Enum\User\RoleUser;
use App\Enum\User\UserStatus;
use App\Models\User;
use Illuminate\Database\Seeder;
use Illuminate\Support\Facades\Hash;

class SuperAdminSeeder extends Seeder
{
    public function run(): void
    {
        User::query()->updateOrCreate(
            [
                'email' => 'superadmin@ff789.club',
            ],
            [
                'name' => 'Super Admin',
                'phone' => '0901000000',
                'password' => Hash::make('password'),
                'role' => RoleUser::SUPER_ADMIN,
                'status' => UserStatus::ACTIVE,
                'email_verified_at' => now(),
                'phone_verified_at' => now(),
            ],
        );
    }
}
