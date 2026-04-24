<?php

namespace App\Models\Content;

use App\Models\User;
use App\Support\Media\WebpImageConverter;
use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;
use Illuminate\Database\Eloquent\SoftDeletes;

class Banner extends Model
{
    use SoftDeletes;

    protected $fillable = [
        'title',
        'image_path',
        'link_url',
        'placement',
        'sort_order',
        'is_active',
        'start_at',
        'end_at',
        'created_by',
    ];

    protected function casts(): array
    {
        return [
            'is_active' => 'boolean',
            'start_at' => 'datetime',
            'end_at' => 'datetime',
        ];
    }

    protected static function booted(): void
    {
        static::saving(function (self $banner): void {
            $banner->image_path = WebpImageConverter::convertPublicDiskPath($banner->image_path);
        });
    }

    public function createdBy(): BelongsTo
    {
        return $this->belongsTo(User::class, 'created_by');
    }
}
