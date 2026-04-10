<?php

namespace App\Filament\Resources;

use Filament\Resources\Resource;
use Illuminate\Support\Facades\Gate;

abstract class BaseResource extends Resource
{
    abstract protected static function abilityPrefix(): string;

    protected static bool $canCreateRecords = true;
    protected static bool $canUpdateRecords = true;
    protected static bool $canDeleteRecords = true;
    protected static bool $canDeleteAnyRecords = true;
    protected static bool $canForceDeleteRecords = true;
    protected static bool $canForceDeleteAnyRecords = true;
    protected static bool $canRestoreRecords = true;
    protected static bool $canRestoreAnyRecords = true;

    public static function shouldRegisterNavigation(): bool
    {
        return false;
    }

    protected static bool $hasTitleCaseModelLabel = false;

    public static function getModelLabel(): string
    {
        if (filled(static::$modelLabel)) {
            return static::$modelLabel;
        }

        if (filled(static::$navigationLabel)) {
            return static::$navigationLabel;
        }

        return parent::getModelLabel();
    }

    public static function getPluralModelLabel(): string
    {
        if (filled(static::$pluralModelLabel)) {
            return static::$pluralModelLabel;
        }

        if (filled(static::$navigationLabel)) {
            return static::$navigationLabel;
        }

        return parent::getPluralModelLabel();
    }

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
        return static::$canCreateRecords && static::gateAllows('create');
    }

    public static function canEdit($record): bool
    {
        return static::$canUpdateRecords && static::gateAllows('update');
    }

    public static function canDelete($record): bool
    {
        return static::$canDeleteRecords && static::gateAllows('delete');
    }

    public static function canDeleteAny(): bool
    {
        return static::$canDeleteAnyRecords && static::gateAllows('deleteAny');
    }

    public static function canForceDelete($record): bool
    {
        return static::$canForceDeleteRecords && static::gateAllows('forceDelete');
    }

    public static function canForceDeleteAny(): bool
    {
        return static::$canForceDeleteAnyRecords && static::gateAllows('forceDeleteAny');
    }

    public static function canRestore($record): bool
    {
        return static::$canRestoreRecords && static::gateAllows('restore');
    }

    public static function canRestoreAny(): bool
    {
        return static::$canRestoreAnyRecords && static::gateAllows('restoreAny');
    }
}
