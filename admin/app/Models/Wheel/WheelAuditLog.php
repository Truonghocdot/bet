<?php

namespace App\Models\Wheel;

use Illuminate\Database\Eloquent\Model;

class WheelAuditLog extends Model
{
    public $timestamps = false;

    protected $fillable = ['campaign_id', 'invitation_id', 'session_id', 'actor_user_id', 'action', 'old_values', 'new_values', 'ip_address', 'created_at'];

    protected function casts(): array
    {
        return ['old_values' => 'array', 'new_values' => 'array', 'created_at' => 'datetime'];
    }
}
