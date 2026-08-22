<?php

namespace App\Console\Commands;

use App\Models\Wheel\WheelInvitation;
use App\Services\Wheel\WheelCampaignService;
use Illuminate\Console\Command;

class SetWheelBotChatCommand extends Command
{
    protected $signature = 'wheel:bot-chat {invitation : Internal invitation ID} {state=on : on or off}';

    protected $description = 'Bật hoặc tắt bot chat cho một lời mời vòng quay đang hoạt động';

    public function handle(WheelCampaignService $service): int
    {
        $state = strtolower(trim((string) $this->argument('state')));
        if (! in_array($state, ['on', 'off'], true)) {
            $this->error('State chỉ nhận on hoặc off.');

            return self::INVALID;
        }

        $invitation = WheelInvitation::query()->findOrFail((int) $this->argument('invitation'));
        $service->setBotChatEnabled($invitation, $state === 'on');
        $this->info(sprintf(
            'Invitation #%d: bot chat %s.',
            $invitation->id,
            $state === 'on' ? 'đã bật' : 'đã tắt',
        ));

        return self::SUCCESS;
    }
}
