<?php

namespace App\Models\Chat;

use App\Models\Wheel\WheelSession;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\Relations\HasMany;

class ChatRoom extends Model
{
    protected $fillable = ['wheel_session_id', 'code', 'name', 'enabled', 'next_bot_at'];

    protected function casts(): array
    {
        return ['enabled' => 'boolean', 'next_bot_at' => 'datetime'];
    }

    public function messages(): HasMany
    {
        return $this->hasMany(ChatMessage::class, 'room_id');
    }

    public function wheelSession(): BelongsTo
    {
        return $this->belongsTo(WheelSession::class, 'wheel_session_id');
    }
}
