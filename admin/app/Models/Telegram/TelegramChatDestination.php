<?php

namespace App\Models\Telegram;

use App\Models\User;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\Relations\HasMany;

class TelegramChatDestination extends Model
{
    protected $fillable = [
        'site_code', 'telegram_chat_id', 'chat_type', 'title', 'username', 'bot_status',
        'is_active', 'discovered_at', 'last_seen_at', 'removed_at', 'last_error', 'last_error_at',
        'activated_by', 'activated_at', 'deactivated_by', 'deactivated_at',
    ];

    protected function casts(): array
    {
        return [
            'telegram_chat_id' => 'integer',
            'is_active' => 'boolean',
            'discovered_at' => 'datetime',
            'last_seen_at' => 'datetime',
            'removed_at' => 'datetime',
            'last_error_at' => 'datetime',
            'activated_at' => 'datetime',
            'deactivated_at' => 'datetime',
        ];
    }

    public function activatedBy(): BelongsTo
    {
        return $this->belongsTo(User::class, 'activated_by');
    }

    public function deactivatedBy(): BelongsTo
    {
        return $this->belongsTo(User::class, 'deactivated_by');
    }

    public function audits(): HasMany
    {
        return $this->hasMany(TelegramChatDestinationAudit::class, 'destination_id');
    }
}
