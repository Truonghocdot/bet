<?php

namespace App\Filament\Resources;

use Filament\Resources\Resource;
use Illuminate\Support\Facades\Gate;

abstract class BaseResource extends Resource
{
    abstract protected static function abilityPrefix(): string;

    protected static function gateAllows(string $ability): bool
    {
        return Gate::allows(static::abilityPrefix().'.'.$ability);
    }

    public static function canViewAny(): bool
    {
        return static::gateAllows('viewAny');
    }

    public static function canView($record): bool
    {
        return static::gateAllows('view');
    }

    public static function canCreate(): bool
    {
        return static::gateAllows('create');
    }

    public static function canEdit($record): bool
    {
        return static::gateAllows('update');
    }

    public static function canDelete($record): bool
    {
        return static::gateAllows('delete');
    }

    public static function canDeleteAny(): bool
    {
        return static::gateAllows('deleteAny');
    }

    public static function canForceDelete($record): bool
    {
        return static::gateAllows('forceDelete');
    }

    public static function canForceDeleteAny(): bool
    {
        return static::gateAllows('forceDeleteAny');
    }

    public static function canRestore($record): bool
    {
        return static::gateAllows('restore');
    }

    public static function canRestoreAny(): bool
    {
        return static::gateAllows('restoreAny');
    }
}
