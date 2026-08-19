<?php

namespace App\Console\Commands;

use App\Services\Wheel\WheelEventPublisher;
use Illuminate\Console\Command;

class PublishWheelOutboxCommand extends Command
{
    protected $signature = 'wheel:publish-outbox {--limit=100}';

    protected $description = 'Phát lại các realtime event vòng quay chưa publish';

    public function handle(WheelEventPublisher $publisher): int
    {
        if (config('wheel.enabled')) {
            $publisher->publishPending(max(1, (int) $this->option('limit')));
        }

        return self::SUCCESS;
    }
}
