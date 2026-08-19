<?php

namespace App\Console\Commands;

use App\Jobs\GenerateChatBotMessage;
use Illuminate\Console\Command;

class GenerateChatBotMessageCommand extends Command
{
    protected $signature = 'chat:generate-bot-message';

    protected $description = 'Đưa một tin nhắn bot chat global vào queue';

    public function handle(): int
    {
        if (! config('wheel.enabled') && ! filter_var(env('CHAT_GLOBAL_ENABLED', false), FILTER_VALIDATE_BOOL)) {
            return self::SUCCESS;
        }

        GenerateChatBotMessage::dispatch()->delay(now()->addSeconds(random_int(0, 19)));

        return self::SUCCESS;
    }
}
