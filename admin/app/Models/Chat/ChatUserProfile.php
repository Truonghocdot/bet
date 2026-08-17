<?php

namespace App\Models\Chat;

use App\Models\User;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

class ChatUserProfile extends Model
{
    protected $fillable = ['user_id', 'display_name'];

    public function user(): BelongsTo
    {
        return $this->belongsTo(User::class);
    }
}
