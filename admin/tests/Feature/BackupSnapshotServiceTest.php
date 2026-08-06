<?php

namespace Tests\Feature;

use App\Services\Backup\BackupSnapshotService;
use Aws\MockHandler;
use Aws\Result;
use Carbon\CarbonImmutable;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Schema;
use Illuminate\Support\Facades\Storage;
use Psr\Http\Message\RequestInterface;
use RuntimeException;
use Tests\TestCase;

class BackupSnapshotServiceTest extends TestCase
{
    protected function setUp(): void
    {
        parent::setUp();

        Schema::dropIfExists('affiliate_reward_logs');
        Schema::dropIfExists('affiliate_referrals');
        Schema::dropIfExists('affiliate_links');
        Schema::dropIfExists('affiliate_profiles');
        Schema::dropIfExists('affiliate_reward_settings');
        Schema::dropIfExists('banners');
        Schema::dropIfExists('news_articles');
        Schema::dropIfExists('wallet_ledger_entries');
        Schema::dropIfExists('transactions');
        Schema::dropIfExists('payment_receiving_accounts');
        Schema::dropIfExists('wallets');
        Schema::dropIfExists('users');
        Schema::create('users', function (Blueprint $table): void {
            $table->id();
            $table->string('name');
            $table->string('email')->nullable()->unique();
            $table->string('password');
            $table->unsignedTinyInteger('role')->default(2);
            $table->unsignedTinyInteger('status')->default(1);
            $table->timestamps();
        });
        Schema::create('wallets', function (Blueprint $table): void {
            $table->id();
            $table->foreignId('user_id')->constrained();
            $table->unsignedTinyInteger('unit');
            $table->decimal('balance', 20, 8)->default(0);
            $table->decimal('locked_balance', 20, 8)->default(0);
            $table->unsignedTinyInteger('status')->default(1);
            $table->timestamps();
        });
        Schema::create('payment_receiving_accounts', function (Blueprint $table): void {
            $table->id();
            $table->string('account_name')->nullable();
            $table->string('account_number')->nullable();
            $table->timestamps();
        });
        Schema::create('transactions', function (Blueprint $table): void {
            $table->id();
            $table->foreignId('user_id')->constrained();
            $table->foreignId('wallet_id')->constrained();
            $table->foreignId('receiving_account_id')->nullable()->constrained('payment_receiving_accounts')->nullOnDelete();
            $table->unsignedTinyInteger('unit');
            $table->unsignedTinyInteger('type');
            $table->decimal('amount', 20, 8);
            $table->decimal('fee', 20, 8)->default(0);
            $table->decimal('net_amount', 20, 8);
            $table->unsignedTinyInteger('status');
            $table->timestamps();
        });
        Schema::create('wallet_ledger_entries', function (Blueprint $table): void {
            $table->id();
            $table->foreignId('wallet_id')->constrained();
            $table->foreignId('user_id')->constrained();
            $table->unsignedTinyInteger('direction');
            $table->decimal('amount', 20, 8);
            $table->decimal('balance_before', 20, 8);
            $table->decimal('balance_after', 20, 8);
            $table->string('reference_type');
            $table->unsignedBigInteger('reference_id')->nullable();
            $table->string('note')->nullable();
            $table->timestamp('created_at');
        });
        Schema::create('affiliate_reward_settings', function (Blueprint $table): void {
            $table->id();
            $table->string('name');
            $table->unsignedInteger('required_qualified_referrals');
            $table->decimal('reward_amount', 20, 8);
            $table->unsignedTinyInteger('unit');
            $table->boolean('is_active')->default(true);
            $table->timestamps();
        });
        Schema::create('affiliate_profiles', function (Blueprint $table): void {
            $table->id();
            $table->foreignId('user_id')->constrained();
            $table->string('ref_code')->unique();
            $table->string('ref_link');
            $table->unsignedTinyInteger('status');
            $table->timestamps();
        });
        Schema::create('affiliate_links', function (Blueprint $table): void {
            $table->id();
            $table->foreignId('affiliate_profile_id')->constrained();
            $table->string('campaign_name');
            $table->string('tracking_code')->unique();
            $table->string('landing_url');
            $table->unsignedTinyInteger('status');
            $table->timestamps();
        });
        Schema::create('affiliate_referrals', function (Blueprint $table): void {
            $table->id();
            $table->foreignId('affiliate_profile_id')->constrained();
            $table->foreignId('referrer_user_id')->constrained('users');
            $table->foreignId('referred_user_id')->unique()->constrained('users');
            $table->foreignId('affiliate_link_id')->nullable()->constrained()->nullOnDelete();
            $table->foreignId('first_deposit_transaction_id')->nullable()->constrained('transactions')->nullOnDelete();
            $table->decimal('first_deposit_amount', 20, 8)->nullable();
            $table->timestamp('qualified_at')->nullable();
            $table->unsignedTinyInteger('status');
            $table->timestamps();
        });
        Schema::create('affiliate_reward_logs', function (Blueprint $table): void {
            $table->id();
            $table->foreignId('affiliate_profile_id')->constrained();
            $table->foreignId('referrer_user_id')->constrained('users');
            $table->foreignId('setting_id')->constrained('affiliate_reward_settings');
            $table->unsignedInteger('required_qualified_referrals');
            $table->unsignedInteger('actual_qualified_referrals');
            $table->decimal('reward_amount', 20, 8);
            $table->unsignedTinyInteger('unit');
            $table->unsignedTinyInteger('status');
            $table->foreignId('wallet_ledger_entry_id')->nullable()->constrained()->nullOnDelete();
            $table->timestamp('granted_at')->nullable();
            $table->timestamps();
        });
        Schema::create('news_articles', function (Blueprint $table): void {
            $table->id();
            $table->string('title');
            $table->string('slug')->unique();
            $table->text('content');
            $table->string('cover_image_path')->nullable();
            $table->boolean('is_published')->default(false);
            $table->foreignId('created_by')->nullable()->constrained('users')->nullOnDelete();
            $table->timestamps();
        });
        Schema::create('banners', function (Blueprint $table): void {
            $table->id();
            $table->string('title')->nullable();
            $table->string('image_path');
            $table->string('placement')->default('home');
            $table->unsignedInteger('sort_order')->default(0);
            $table->boolean('is_active')->default(true);
            $table->foreignId('created_by')->nullable()->constrained('users')->nullOnDelete();
            $table->timestamps();
        });

        Storage::fake('r2');
        Storage::fake('public');
        config()->set('backup.disk', 'r2');
        config()->set('backup.site', 'test-site');
        config()->set('backup.prefix', 'backups');
        config()->set('backup.asset_disk', 'public');
    }

    public function test_it_creates_verifies_and_restores_a_snapshot(): void
    {
        DB::table('users')->insert([
            [
                'id' => 123456,
                'name' => 'Original Referrer',
                'email' => 'referrer@example.com',
                'password' => '$2y$04$test-hash',
                'role' => 2,
                'status' => 1,
                'created_at' => now(),
                'updated_at' => now(),
            ],
            [
                'id' => 123457,
                'name' => 'Original Referred User',
                'email' => 'referred@example.com',
                'password' => '$2y$04$test-hash',
                'role' => 2,
                'status' => 1,
                'created_at' => now(),
                'updated_at' => now(),
            ],
        ]);
        DB::table('wallets')->insert([
            [
                'id' => 91,
                'user_id' => 123456,
                'unit' => 1,
                'balance' => '123456789.12345678',
                'locked_balance' => '200.00000000',
                'status' => 1,
                'created_at' => now(),
                'updated_at' => now(),
            ],
            [
                'id' => 90,
                'user_id' => 123457,
                'unit' => 1,
                'balance' => '500.00000000',
                'locked_balance' => '0.00000000',
                'status' => 1,
                'created_at' => now(),
                'updated_at' => now(),
            ],
        ]);
        DB::table('payment_receiving_accounts')->insert([
            'id' => 80,
            'account_name' => 'Backup Bank',
            'account_number' => '0123456789',
            'created_at' => now(),
            'updated_at' => now(),
        ]);
        DB::table('transactions')->insert([
            'id' => 94,
            'user_id' => 123457,
            'wallet_id' => 90,
            'receiving_account_id' => 80,
            'unit' => 1,
            'type' => 1,
            'amount' => '500.00000000',
            'fee' => '0.00000000',
            'net_amount' => '500.00000000',
            'status' => 2,
            'created_at' => now(),
            'updated_at' => now(),
        ]);
        DB::table('wallet_ledger_entries')->insert([
            'id' => 92,
            'wallet_id' => 91,
            'user_id' => 123456,
            'direction' => 1,
            'amount' => '123456789.12345678',
            'balance_before' => '0.00000000',
            'balance_after' => '123456789.12345678',
            'reference_type' => 'backup-test',
            'created_at' => now(),
        ]);
        DB::table('affiliate_reward_settings')->insert([
            'id' => 101,
            'name' => 'First qualified referral',
            'required_qualified_referrals' => 1,
            'reward_amount' => '100.00000000',
            'unit' => 1,
            'is_active' => true,
            'created_at' => now(),
            'updated_at' => now(),
        ]);
        DB::table('affiliate_profiles')->insert([
            'id' => 102,
            'user_id' => 123456,
            'ref_code' => 'REFBACKUP',
            'ref_link' => 'https://example.test/register?ref_code=REFBACKUP',
            'status' => 1,
            'created_at' => now(),
            'updated_at' => now(),
        ]);
        DB::table('affiliate_links')->insert([
            'id' => 103,
            'affiliate_profile_id' => 102,
            'campaign_name' => 'Backup campaign',
            'tracking_code' => 'BACKUP-CAMPAIGN',
            'landing_url' => 'https://example.test/register',
            'status' => 1,
            'created_at' => now(),
            'updated_at' => now(),
        ]);
        DB::table('affiliate_referrals')->insert([
            'id' => 104,
            'affiliate_profile_id' => 102,
            'referrer_user_id' => 123456,
            'referred_user_id' => 123457,
            'affiliate_link_id' => 103,
            'first_deposit_transaction_id' => 94,
            'first_deposit_amount' => '500.00000000',
            'qualified_at' => now(),
            'status' => 2,
            'created_at' => now(),
            'updated_at' => now(),
        ]);
        DB::table('affiliate_reward_logs')->insert([
            'id' => 105,
            'affiliate_profile_id' => 102,
            'referrer_user_id' => 123456,
            'setting_id' => 101,
            'required_qualified_referrals' => 1,
            'actual_qualified_referrals' => 1,
            'reward_amount' => '100.00000000',
            'unit' => 1,
            'status' => 2,
            'wallet_ledger_entry_id' => 92,
            'granted_at' => now(),
            'created_at' => now(),
            'updated_at' => now(),
        ]);
        DB::table('news_articles')->insert([
            'id' => 93,
            'title' => 'Original News',
            'slug' => 'original-news',
            'content' => '<p>Body</p>',
            'cover_image_path' => 'news/cover.webp',
            'is_published' => true,
            'created_by' => 123456,
            'created_at' => now(),
            'updated_at' => now(),
        ]);
        DB::table('banners')->insert([
            'id' => 106,
            'title' => 'Original Banner',
            'image_path' => 'banners/home.webp',
            'placement' => 'home',
            'sort_order' => 1,
            'is_active' => true,
            'created_by' => 123456,
            'created_at' => now(),
            'updated_at' => now(),
        ]);
        Storage::disk('public')->put('news/cover.webp', 'cover-image');
        Storage::disk('public')->put('news/editor/inline.webp', 'inline-image');
        Storage::disk('public')->put('banners/home.webp', 'banner-image');

        $service = app(BackupSnapshotService::class);
        $manifest = $service->create();
        $snapshot = $manifest['snapshot'];

        Storage::disk('r2')->assertExists("backups/test-site/_manifests/{$snapshot}.json");
        $this->assertSame($snapshot, $service->all()[0]['snapshot']);
        $verified = $service->verify($snapshot);
        $this->assertSame(15, $verified['objects']);

        DB::table('users')->where('id', 123456)->update(['name' => 'Changed Referrer']);
        DB::table('wallets')->where('id', 91)->update(['balance' => '1.00000000']);
        DB::table('transactions')->where('id', 94)->update(['amount' => '1.00000000']);
        DB::table('affiliate_profiles')->where('id', 102)->update(['ref_code' => 'CHANGED']);
        DB::table('affiliate_referrals')->where('id', 104)->update(['status' => 1]);
        DB::table('news_articles')->where('id', 93)->update(['title' => 'Changed News']);
        DB::table('banners')->where('id', 106)->update(['title' => 'Changed Banner']);
        Storage::disk('public')->delete('news/cover.webp');
        Storage::disk('public')->delete('banners/home.webp');

        $restored = $service->restore($snapshot);

        $this->assertSame(14, $restored['rows']);
        $this->assertSame(3, $restored['assets']);
        $this->assertDatabaseHas('users', ['id' => 123456, 'name' => 'Original Referrer']);
        $this->assertDatabaseHas('wallets', ['id' => 91, 'balance' => 123456789.12345678]);
        $this->assertDatabaseHas('transactions', ['id' => 94, 'amount' => 500]);
        $this->assertDatabaseHas('affiliate_profiles', ['id' => 102, 'user_id' => 123456, 'ref_code' => 'REFBACKUP']);
        $this->assertDatabaseHas('affiliate_referrals', [
            'id' => 104,
            'referrer_user_id' => 123456,
            'referred_user_id' => 123457,
            'first_deposit_transaction_id' => 94,
            'status' => 2,
        ]);
        $this->assertDatabaseHas('affiliate_reward_logs', [
            'id' => 105,
            'affiliate_profile_id' => 102,
            'wallet_ledger_entry_id' => 92,
        ]);
        $this->assertDatabaseHas('news_articles', ['id' => 93, 'title' => 'Original News']);
        $this->assertDatabaseHas('banners', ['id' => 106, 'title' => 'Original Banner']);
        Storage::disk('public')->assertExists('news/cover.webp');
        Storage::disk('public')->assertExists('banners/home.webp');
        $this->assertSame('inline-image', Storage::disk('public')->get('news/editor/inline.webp'));
    }

    public function test_verify_rejects_a_corrupted_object(): void
    {
        $service = app(BackupSnapshotService::class);
        $manifest = $service->create();
        $snapshot = $manifest['snapshot'];
        $object = $manifest['database']['tables'][0]['path'];

        Storage::disk('r2')->put("backups/test-site/{$snapshot}/{$object}", 'corrupted');

        $this->expectException(RuntimeException::class);
        $this->expectExceptionMessage('checksum mismatch');

        $service->verify($snapshot);
    }

    public function test_v1_snapshots_remain_listable_and_verifiable(): void
    {
        $service = app(BackupSnapshotService::class);
        $manifest = $service->create();
        $snapshot = $manifest['snapshot'];
        $legacyTables = ['users', 'wallets', 'wallet_ledger_entries', 'news_articles'];
        $manifest['format'] = 'site-backup-v1';
        $manifest['database']['tables'] = array_values(array_filter(
            $manifest['database']['tables'],
            static fn (array $table): bool => in_array($table['name'], $legacyTables, true),
        ));

        Storage::disk('r2')->put(
            "backups/test-site/_manifests/{$snapshot}.json",
            json_encode($manifest, JSON_THROW_ON_ERROR),
        );

        $this->assertSame($snapshot, $service->all()[0]['snapshot']);
        $this->assertSame(4, $service->verify($snapshot)['objects']);
    }

    public function test_r2_uploads_do_not_send_unsupported_acl_headers(): void
    {
        $requests = [];
        $responses = [];

        for ($index = 0; $index < 13; $index++) {
            $responses[] = static function ($command, RequestInterface $request) use (&$requests): Result {
                $requests[] = $request;

                return new Result(['ETag' => 'test-etag']);
            };
        }

        config()->set('filesystems.disks.r2-test', [
            'driver' => 's3',
            'key' => 'test-key',
            'secret' => 'test-secret',
            'region' => 'auto',
            'bucket' => 'test-bucket',
            'endpoint' => 'https://account.r2.cloudflarestorage.com',
            'handler' => new MockHandler($responses),
            'throw' => true,
        ]);

        app(BackupSnapshotService::class)->create('r2-test');

        $this->assertCount(13, $requests);
        foreach ($requests as $request) {
            $this->assertFalse($request->hasHeader('x-amz-acl'));
        }
    }

    public function test_prune_removes_expired_objects_and_the_catalog_manifest(): void
    {
        CarbonImmutable::setTestNow('2026-08-01 00:00:00 UTC');
        config()->set('backup.retention_days', 1);
        config()->set('backup.keep_minimum', 0);

        try {
            $service = app(BackupSnapshotService::class);
            $snapshot = $service->create()['snapshot'];
            CarbonImmutable::setTestNow('2026-08-03 00:00:00 UTC');

            $this->assertSame([$snapshot], $service->prune());
            Storage::disk('r2')->assertMissing("backups/test-site/_manifests/{$snapshot}.json");
            Storage::disk('r2')->assertMissing("backups/test-site/{$snapshot}/database/users.jsonl.gz");
        } finally {
            CarbonImmutable::setTestNow();
        }
    }
}
