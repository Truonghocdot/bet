<?php

namespace App\Models\Wheel;

use App\Models\User;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\Relations\HasMany;

class WheelCampaign extends Model
{
    protected $fillable = ['name', 'status', 'opens_at', 'closes_at', 'duration_seconds', 'spin_duration_seconds', 'created_by'];

    protected function casts(): array
    {
        return ['opens_at' => 'datetime', 'closes_at' => 'datetime', 'duration_seconds' => 'integer', 'spin_duration_seconds' => 'integer'];
    }

    public function roundTemplates(): HasMany
    {
        return $this->hasMany(WheelCampaignRoundTemplate::class, 'campaign_id')->orderBy('round_no');
    }

    public function invitations(): HasMany
    {
        return $this->hasMany(WheelInvitation::class, 'campaign_id');
    }

    public function createdBy(): BelongsTo
    {
        return $this->belongsTo(User::class, 'created_by');
    }
}
