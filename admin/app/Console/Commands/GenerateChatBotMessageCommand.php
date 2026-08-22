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

        if (config('wheel.enabled') || filter_var(env('CHAT_GLOBAL_ENABLED', false), FILTER_VALIDATE_BOOL)) {
            // Each event room schedules its next job after a message is
            // created. This generic dispatch is only a recovery sweep for
            // rooms whose delayed job was lost or never created; it also
            // handles the global room when that feature is enabled.
            GenerateChatBotMessage::dispatch();
        }

        return self::SUCCESS;
    }
}
