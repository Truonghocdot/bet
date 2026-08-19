<?php

namespace App\Models\Wheel;

use App\Models\Chat\ChatRoom;
use App\Models\User;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\Relations\HasMany;
use Illuminate\Database\Eloquent\Relations\HasOne;

class WheelSession extends Model
{
    protected $fillable = ['public_id', 'invitation_id', 'user_id', 'status', 'current_round', 'started_at', 'ends_at', 'completed_at', 'expired_at', 'version'];

    protected function casts(): array
    {
        return ['current_round' => 'integer', 'started_at' => 'datetime', 'ends_at' => 'datetime', 'completed_at' => 'datetime', 'expired_at' => 'datetime', 'version' => 'integer'];
    }

    public function invitation(): BelongsTo
    {
        return $this->belongsTo(WheelInvitation::class, 'invitation_id');
    }

    public function user(): BelongsTo
    {
        return $this->belongsTo(User::class);
    }

    public function rewards(): HasMany
    {
        return $this->hasMany(WheelReward::class, 'session_id');
    }

    public function chatRoom(): HasOne
    {
        return $this->hasOne(ChatRoom::class, 'wheel_session_id');
    }
}
