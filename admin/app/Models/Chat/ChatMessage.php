<?php

namespace App\Models\Chat;

use App\Models\User;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

class ChatMessage extends Model
{
    public const STATUS_VISIBLE = 1;

    public const STATUS_HIDDEN = 2;

    public const STATUS_DELETED = 3;

    protected $fillable = [
        'room_id', 'actor_type', 'user_id', 'bot_profile_id', 'display_name', 'body', 'status',
        'moderated_by', 'moderated_at',
    ];

    protected function casts(): array
    {
        return ['status' => 'integer', 'moderated_at' => 'datetime'];
    }

    public function room(): BelongsTo
    {
        return $this->belongsTo(ChatRoom::class, 'room_id');
    }

    public function user(): BelongsTo
    {
        return $this->belongsTo(User::class);
    }

    public function botProfile(): BelongsTo
    {
        return $this->belongsTo(ChatBotProfile::class, 'bot_profile_id');
    }

    public function moderatedBy(): BelongsTo
    {
        return $this->belongsTo(User::class, 'moderated_by');
    }
}
