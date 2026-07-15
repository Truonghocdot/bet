<?php

namespace App\Support\Media;

use Illuminate\Support\Facades\Log;
use Illuminate\Support\Facades\Storage;

class WebpImageConverter
{
    public static function ensurePublicDiskWebpVariant(?string $path, int $quality = 82): ?string
    {
        $path = trim((string) $path);
        if ($path === '') {
            return null;
        }

        if (str_ends_with(strtolower($path), '.webp')) {
            return $path;
        }

        $disk = Storage::disk('public');
        $dirname = pathinfo($path, PATHINFO_DIRNAME);
        $filename = pathinfo($path, PATHINFO_FILENAME);
        $webpPath = ($dirname !== '' && $dirname !== '.' ? $dirname.'/' : '').$filename.'.webp';

        if ($disk->exists($webpPath)) {
            return $webpPath;
        }

        if (! function_exists('imagecreatefromstring') || ! function_exists('imagewebp')) {
            Log::warning('[media.webp] gd extension is unavailable, skip sidecar generation', ['path' => $path]);

            return null;
        }

        if (! $disk->exists($path)) {
            return null;
        }

        try {
            $binary = $disk->get($path);
            $image = @imagecreatefromstring($binary);
            if ($image === false) {
                return null;
            }

            if (function_exists('imagepalettetotruecolor')) {
                @imagepalettetotruecolor($image);
            }
            if (function_exists('imagealphablending')) {
                @imagealphablending($image, true);
            }
            if (function_exists('imagesavealpha')) {
                @imagesavealpha($image, true);
            }

            ob_start();
            $ok = imagewebp($image, null, $quality);
            $webpBinary = ob_get_clean();
            imagedestroy($image);

            if (! $ok || ! is_string($webpBinary) || $webpBinary === '') {
                return null;
            }

            $disk->put($webpPath, $webpBinary);

            return $webpPath;
        } catch (\Throwable $e) {
            Log::warning('[media.webp] sidecar generation failed', ['path' => $path, 'error' => $e->getMessage()]);

            return null;
        }
    }

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
            if (function_exists('imagealphablending')) {
                @imagealphablending($image, true);
            }
            if (function_exists('imagesavealpha')) {
                @imagesavealpha($image, true);
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
