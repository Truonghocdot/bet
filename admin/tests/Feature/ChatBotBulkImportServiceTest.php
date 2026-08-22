<?php

namespace Tests\Feature;

use App\Services\Chat\ChatBotBulkImportService;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Schema;
use Tests\TestCase;

class ChatBotBulkImportServiceTest extends TestCase
{
    protected function setUp(): void
    {
        parent::setUp();

        Schema::create('chat_bot_profiles', function (Blueprint $table): void {
            $table->id();
            $table->string('display_name');
            $table->string('avatar_path')->nullable();
            $table->boolean('active')->default(true);
            $table->unsignedInteger('sort_order')->default(0);
            $table->timestamps();
        });
        Schema::create('chat_bot_templates', function (Blueprint $table): void {
            $table->id();
            $table->unsignedBigInteger('bot_profile_id')->nullable();
            $table->string('body', 280);
            $table->string('category')->default('general');
            $table->string('language')->default('vi');
            $table->boolean('active')->default(true);
            $table->timestamp('last_used_at')->nullable();
            $table->unsignedInteger('usage_count')->default(0);
            $table->timestamps();
        });
    }

    protected function tearDown(): void
    {
        Schema::dropIfExists('chat_bot_templates');
        Schema::dropIfExists('chat_bot_profiles');

        parent::tearDown();
    }

    public function test_it_imports_the_customer_tsv_format_and_can_be_run_again(): void
    {
        DB::table('chat_bot_profiles')->insert([
            'display_name' => 'Tám Lì',
            'active' => true,
            'sort_order' => 1,
            'created_at' => now(),
            'updated_at' => now(),
        ]);
        DB::table('chat_bot_templates')->insert([
            'body' => 'Câu mẫu cũ không còn dùng',
            'category' => 'general',
            'language' => 'vi',
            'active' => true,
            'usage_count' => 0,
            'created_at' => now(),
            'updated_at' => now(),
        ]);
        $input = implode("\n", [
            "người chơi# 258425\tNhìn thôi cũng muốn thử vận may",
            "người chơi# 195658\t🥰🥰🥰",
            "người chơi# 258425\tChúc ai nhận được quà sẽ tận hưởng nhé",
            "người chơi# 999999\thttps://example.com không hợp lệ",
        ]);

        $first = app(ChatBotBulkImportService::class)->import($input, replaceExistingTemplates: true);

        self::assertSame(3, $first['templates_created']);
        self::assertSame(2, $first['profiles_created']);
        self::assertSame(1, $first['invalid']);
        self::assertTrue($first['cleanup_skipped']);
        self::assertSame(0, $first['templates_deactivated']);
        self::assertSame(1, $first['profiles_deactivated']);
        self::assertDatabaseHas('chat_bot_profiles', ['display_name' => 'ID game #258425', 'active' => true]);
        self::assertDatabaseHas('chat_bot_profiles', ['display_name' => 'ID game #195658', 'active' => true]);
        self::assertDatabaseHas('chat_bot_profiles', ['display_name' => 'Tám Lì', 'active' => false]);
        self::assertSame(2, DB::table('chat_bot_templates')
            ->join('chat_bot_profiles', 'chat_bot_profiles.id', '=', 'chat_bot_templates.bot_profile_id')
            ->where('chat_bot_profiles.display_name', 'ID game #258425')
            ->count());

        $cleanInput = implode("\n", array_slice(explode("\n", $input), 0, 3));
        $second = app(ChatBotBulkImportService::class)->import($cleanInput, replaceExistingTemplates: true);

        self::assertSame(0, $second['templates_created']);
        self::assertSame(3, $second['duplicates']);
        self::assertSame(1, $second['templates_deactivated']);
        self::assertFalse($second['cleanup_skipped']);
        self::assertDatabaseHas('chat_bot_templates', ['body' => 'Câu mẫu cũ không còn dùng', 'active' => false]);
        self::assertSame(4, DB::table('chat_bot_templates')->count());
    }
}
