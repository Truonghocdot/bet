<?php

namespace App\Models\Wheel;

use App\Models\User;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\Relations\HasMany;
use Illuminate\Database\Eloquent\Relations\HasOne;
use Illuminate\Support\Str;

class WheelInvitation extends Model
{
    protected $fillable = ['public_id', 'campaign_id', 'user_id', 'status', 'activated_at', 'expires_at', 'popup_seen_at', 'revoked_at', 'activated_by', 'revoked_by'];

    protected static function booted(): void
    {
        static::creating(function (self $model): void {
            $model->public_id ??= (string) Str::uuid();
        });
    }

    protected function casts(): array
    {
        return ['activated_at' => 'datetime', 'expires_at' => 'datetime', 'popup_seen_at' => 'datetime', 'revoked_at' => 'datetime'];
    }

    public function campaign(): BelongsTo
    {
        return $this->belongsTo(WheelCampaign::class, 'campaign_id');
    }

    public function user(): BelongsTo
    {
        return $this->belongsTo(User::class);
    }

    public function rounds(): HasMany
    {
        return $this->hasMany(WheelInvitationRound::class, 'invitation_id')->orderBy('round_no');
    }

    public function session(): HasOne
    {
        return $this->hasOne(WheelSession::class, 'invitation_id');
    }
}
