<?php

namespace App\Support\Media;

use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Storage;

class WebpImageConverter
{
    public static function convertPublicDiskPath(?string $path, int $quality = 82): ?string
    {
        $path = trim((string) $path);
        if ($path === '' || str_ends_with(strtolower($path), '.webp')) {
            return $path !== '' ? $path : null;
        }

        if (! function_exists('imagecreatefromstring') || ! function_exists('imagewebp')) {
            Log::warning('[media.webp] gd extension is unavailable, skip conversion', ['path' => $path]);
            return $path;
        }

        $disk = Storage::disk('public');
        if (! $disk->exists($path)) {
            return $path;
        }

        try {
            $binary = $disk->get($path);
            $image = @imagecreatefromstring($binary);
            if ($image === false) {
                return $path;
            }

            if (function_exists('imagepalettetotruecolor')) {
                @imagepalettetotruecolor($image);
            }

            $dirname = pathinfo($path, PATHINFO_DIRNAME);
            $filename = pathinfo($path, PATHINFO_FILENAME);
            $webpPath = ($dirname !== '' && $dirname !== '.' ? $dirname.'/' : '').$filename.'.webp';

            ob_start();
            $ok = imagewebp($image, null, $quality);
            $webpBinary = ob_get_clean();
            imagedestroy($image);

            if (! $ok || ! is_string($webpBinary) || $webpBinary === '') {
                return $path;
            }

            $disk->put($webpPath, $webpBinary);
            if ($webpPath !== $path) {
                $disk->delete($path);
            }

            return $webpPath;
        } catch (\Throwable $e) {
            Log::warning('[media.webp] conversion failed', ['path' => $path, 'error' => $e->getMessage()]);
            return $path;
        }
    }
}

