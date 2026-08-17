<?php

namespace App\Models\Chat;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\HasMany;

class ChatBotProfile extends Model
{
    protected $fillable = ['display_name', 'avatar_path', 'active', 'sort_order'];

    protected function casts(): array
    {
        return ['active' => 'boolean'];
    }

    public function templates(): HasMany
    {
        return $this->hasMany(ChatBotTemplate::class, 'bot_profile_id');
    }

    public function messages(): HasMany
    {
        return $this->hasMany(ChatMessage::class, 'bot_profile_id');
    }
}
