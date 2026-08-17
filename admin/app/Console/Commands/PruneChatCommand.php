<?php

namespace App\Console\Commands;

use Illuminate\Console\Command;
use Illuminate\Support\Facades\DB;

class PruneChatCommand extends Command
{
    protected $signature = 'chat:prune {--message-hours=6} {--audit-days=90}';

    protected $description = 'Xóa tin chat và audit quá thời hạn lưu trữ';

    public function handle(): int
    {
        $messageHours = max(1, (int) $this->option('message-hours'));
        $messages = DB::table('chat_messages')->where('created_at', '<', now()->subHours($messageHours))->delete();
        $audits = DB::table('chat_moderation_actions')->where('created_at', '<', now()->subDays((int) $this->option('audit-days')))->delete();
        $this->info("Đã xóa {$messages} tin nhắn và {$audits} bản ghi audit.");

        return self::SUCCESS;
    }
}
