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

        // The room controls its own 8-14 second jitter. Dispatch immediately
        // so a short-lived event session is not delayed by another scheduler
        // jitter window.
        GenerateChatBotMessage::dispatch();

        return self::SUCCESS;
    }
}
