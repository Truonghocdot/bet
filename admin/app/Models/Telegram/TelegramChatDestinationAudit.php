<?php

namespace App\Models\Telegram;

use App\Models\User;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

class TelegramChatDestinationAudit extends Model
{
    public $timestamps = false;

    protected $fillable = ['destination_id', 'actor_user_id', 'action', 'old_values', 'new_values', 'ip_address', 'created_at'];

    protected function casts(): array
    {
        return ['old_values' => 'array', 'new_values' => 'array', 'created_at' => 'datetime'];
    }

    public function destination(): BelongsTo
    {
        return $this->belongsTo(TelegramChatDestination::class, 'destination_id');
    }

    public function actor(): BelongsTo
    {
        return $this->belongsTo(User::class, 'actor_user_id');
    }
}
