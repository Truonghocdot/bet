<?php

namespace App\Models\Chat;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Validation\ValidationException;

class ChatBotTemplate extends Model
{
    protected $fillable = [
        'bot_profile_id', 'body', 'category', 'language', 'active', 'last_used_at', 'usage_count',
    ];

    protected function casts(): array
    {
        return [
            'active' => 'boolean',
            'last_used_at' => 'datetime',
            'usage_count' => 'integer',
        ];
    }

    protected static function booted(): void
    {
        static::saving(function (self $template): void {
            $body = preg_replace('/\s+/u', ' ', trim((string) $template->body)) ?? '';
            $containsUrl = preg_match('/(https?:\/\/|www\.|(?:^|[\s(])(?:[a-z0-9-]+\.)+[a-z]{2,}(?:[\/?#:)]|\s|$))/iu', $body) === 1;

            if ($body === '' || mb_strlen($body) > 280 || $containsUrl) {
                throw ValidationException::withMessages([
                    'body' => 'Câu mẫu phải là text tối đa 280 ký tự và không được chứa URL.',
                ]);
            }

            $template->body = $body;
        });
    }

    public function botProfile(): BelongsTo
    {
        return $this->belongsTo(ChatBotProfile::class, 'bot_profile_id');
    }
}
