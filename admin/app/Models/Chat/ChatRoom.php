<?php

namespace App\Models\Chat;

use App\Models\Wheel\WheelInvitation;
use App\Models\Wheel\WheelSession;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\Relations\HasMany;

class ChatRoom extends Model
{
    protected $fillable = ['wheel_session_id', 'wheel_invitation_id', 'code', 'name', 'enabled', 'next_bot_at', 'bot_active_until', 'bot_message_count'];

    protected function casts(): array
    {
        return ['enabled' => 'boolean', 'next_bot_at' => 'datetime', 'bot_active_until' => 'datetime', 'bot_message_count' => 'integer'];
    }

    public function messages(): HasMany
    {
        return $this->hasMany(ChatMessage::class, 'room_id');
    }

    public function wheelSession(): BelongsTo
    {
        return $this->belongsTo(WheelSession::class, 'wheel_session_id');
    }

    public function wheelInvitation(): BelongsTo
    {
        return $this->belongsTo(WheelInvitation::class, 'wheel_invitation_id');
    }
}
