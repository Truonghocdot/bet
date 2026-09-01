<?php

namespace App\Services\Telegram;

use App\Models\Telegram\TelegramChatDestination;
use App\Models\Telegram\TelegramChatDestinationAudit;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Http;
use Illuminate\Validation\ValidationException;

class TelegramChatDestinationService
{
    public function activate(TelegramChatDestination $destination): void
    {
        $this->setActive($destination, true);
    }

    public function deactivate(TelegramChatDestination $destination): void
    {
        $this->setActive($destination, false);
    }

    public function sendTest(TelegramChatDestination $destination): void
    {
        $siteCode = $this->siteCode();
        if ($destination->site_code !== $siteCode) {
            throw ValidationException::withMessages(['destination' => 'Group không thuộc site hiện tại.']);
        }

        $response = Http::timeout((int) config('services.telegram.timeout', 10))
            ->withHeaders(['X-Internal-Token' => (string) config('services.telegram.gate_internal_token')])
            ->post(rtrim((string) config('services.telegram.gate_base_url'), '/').'/internal/v1/telegram/test', [
                'site_code' => $siteCode,
                'chat_id' => (int) $destination->telegram_chat_id,
                'message' => 'Tin nhắn kiểm tra Telegram '.$siteCode.' lúc '.now()->format('d/m/Y H:i:s'),
            ]);

        if (! $response->successful()) {
            throw ValidationException::withMessages(['destination' => 'Gate không gửi được tin thử: '.$response->body()]);
        }
    }

    private function setActive(TelegramChatDestination $destination, bool $active): void
    {
        if ($destination->site_code !== $this->siteCode()) {
            throw ValidationException::withMessages(['destination' => 'Group không thuộc site hiện tại.']);
        }
        if ($active && ! in_array($destination->bot_status, ['member', 'administrator'], true)) {
            throw ValidationException::withMessages(['destination' => 'Bot không còn quyền thành viên trong group này.']);
        }

        DB::transaction(function () use ($destination, $active): void {
            $locked = TelegramChatDestination::query()->lockForUpdate()->findOrFail($destination->getKey());
            if ($active && ! in_array($locked->bot_status, ['member', 'administrator'], true)) {
                throw ValidationException::withMessages(['destination' => 'Bot không còn quyền thành viên trong group này.']);
            }
            $old = ['is_active' => (bool) $locked->is_active, 'bot_status' => $locked->bot_status];
            $now = now();
            $locked->forceFill([
                'is_active' => $active,
                'activated_by' => $active ? auth()->id() : $locked->activated_by,
                'activated_at' => $active ? $now : $locked->activated_at,
                'deactivated_by' => $active ? null : auth()->id(),
                'deactivated_at' => $active ? null : $now,
                'last_error' => null,
                'last_error_at' => null,
            ])->save();

            TelegramChatDestinationAudit::query()->create([
                'destination_id' => $locked->getKey(),
                'actor_user_id' => auth()->id(),
                'action' => $active ? 'activated' : 'deactivated',
                'old_values' => $old,
                'new_values' => ['is_active' => $active, 'bot_status' => $locked->bot_status],
                'ip_address' => request()->ip(),
                'created_at' => $now,
            ]);
        });

        $destination->refresh();
    }

    private function siteCode(): string
    {
        return trim((string) config('services.telegram.site_code', env('WHEEL_SITE_CODE', 'fh88u')));
    }
}
