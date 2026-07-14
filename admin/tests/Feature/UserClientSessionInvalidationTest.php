<?php

namespace Tests\Feature;

use App\Enum\User\RoleUser;
use App\Enum\User\UserStatus;
use App\Models\User;
use Illuminate\Support\Facades\Artisan;
use Tests\TestCase;

class UserClientSessionInvalidationTest extends TestCase
{
    protected function setUp(): void
    {
        parent::setUp();

        $this->migrateAuthTables();
    }

    public function test_locking_user_revokes_refresh_tokens(): void
    {
        $user = User::factory()->create([
            'role' => RoleUser::CLIENT,
            'status' => UserStatus::ACTIVE,
        ]);

        \DB::table('auth_refresh_tokens')->insert([
            'user_id' => $user->id,
            'token' => 'main:test-refresh-token',
            'expires_at' => now()->addDay(),
            'ip' => '127.0.0.1',
            'user_agent' => 'PHPUnit',
            'created_at' => now(),
            'updated_at' => now(),
        ]);

        $this->assertDatabaseCount('auth_refresh_tokens', 1);

        $user->update([
            'status' => UserStatus::BANNED,
        ]);

        $this->assertDatabaseCount('auth_refresh_tokens', 0);
    }

    private function migrateAuthTables(): void
    {
        Artisan::call('db:wipe', ['--force' => true]);

        foreach ($this->migrationPaths() as $path) {
            Artisan::call('migrate', [
                '--force' => true,
                '--path' => $path,
                '--realpath' => true,
            ]);
        }
    }

    /**
     * @return list<string>
     */
    private function migrationPaths(): array
    {
        return [
            base_path('database/migrations/0001_01_01_000000_create_users_table.php'),
            base_path('database/migrations/2026_04_13_145037_create_auth_refresh_tokens_table.php'),
        ];
    }
}
