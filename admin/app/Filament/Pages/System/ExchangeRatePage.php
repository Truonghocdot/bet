<?php

namespace App\Filament\Pages\System;

use App\Filament\Pages\System\Schemas\ExchangeRatePageForm;
use App\Services\Admin\ExchangeRateService;
use BackedEnum;
use Filament\Actions\Action;
use Filament\Notifications\Notification;
use Filament\Pages\Concerns;
use Filament\Pages\Page;
use Filament\Schemas\Components\Actions;
use Filament\Schemas\Components\Component;
use Filament\Schemas\Components\EmbeddedSchema;
use Filament\Schemas\Components\Form;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Illuminate\Contracts\Support\Htmlable;
use Illuminate\Support\Facades\Gate;
use UnitEnum;

class ExchangeRatePage extends Page
{
    use Concerns\HasMaxWidth;
    use Concerns\HasTopbar;

    protected static bool $isDiscovered = true;

    protected static ?string $slug = 'settings/exchange-rate-usdt-vnd';

    protected static UnitEnum|string|null $navigationGroup = 'Thiết lập';

    protected static ?string $navigationLabel = 'Thiết lập';

    protected static ?string $title = 'Thiết lập';

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedCurrencyDollar;

    protected static ?int $navigationSort = 5;

    public ?array $data = [];

    public function mount(): void
    {
        $this->fillForm();
    }

    public static function canAccess(): bool
    {
        return Gate::allows('system.exchange-rate-settings.manage');
    }

    public function getTitle(): string | Htmlable
    {
        return static::$title;
    }

    public function getHeading(): string | Htmlable | null
    {
        return static::$title;
    }

    public function getSubheading(): string | Htmlable | null
    {
        $snapshot = app(ExchangeRateService::class)->getSnapshot();

        return sprintf(
            'Redis chia sẻ Gin (%s): %s | Cache Laravel (%s): %s',
            $snapshot['redis_connection'],
            $snapshot['redis_key'],
            $snapshot['cache_store'],
            $snapshot['cache_key'],
        );
    }

    public function getLayoutData(): array
    {
        return [
            'hasTopbar' => $this->hasTopbar(),
            'maxContentWidth' => $maxContentWidth = $this->getMaxWidth() ?? $this->getMaxContentWidth(),
            'maxWidth' => $maxContentWidth,
        ];
    }

    public function defaultForm(Schema $schema): Schema
    {
        return $schema->statePath('data');
    }

    public function form(Schema $schema): Schema
    {
        return ExchangeRatePageForm::configure($schema);
    }

    public function content(Schema $schema): Schema
    {
        return $schema->components([
            $this->getFormContentComponent(),
        ]);
    }

    public function getFormContentComponent(): Component
    {
        return Form::make([EmbeddedSchema::make('form')])
            ->id('form')
            ->livewireSubmitHandler('save')
            ->footer([
                Actions::make($this->getFormActions())
                    ->alignment($this->getFormActionsAlignment())
                    ->fullWidth($this->hasFullWidthFormActions())
                    ->sticky($this->areFormActionsSticky())
                    ->key('form-actions'),
            ]);
    }

    protected function getFormActions(): array
    {
        return [
            Action::make('save')
                ->label('Lưu thay đổi')
                ->icon('heroicon-m-check')
                ->submit('save')
                ->color('primary'),
        ];
    }

    protected function hasFullWidthFormActions(): bool
    {
        return true;
    }

    protected function fillForm(): void
    {
        $this->data = app(ExchangeRateService::class)->getSnapshot();
        $this->form->fill($this->data);
    }

    public function save(): void
    {
        $setting = app(ExchangeRateService::class)->saveSetting(
            $this->form->getState(),
            auth()->user(),
        );

        $this->data = app(ExchangeRateService::class)->getSnapshot();
        $this->form->fill($this->data);

        Notification::make()
            ->title(sprintf(
                'Đã cập nhật cấu hình hệ thống. Tỉ giá USDT/VND: %s VND',
                number_format((float) $setting->rate, 0, ',', '.'),
            ))
            ->success()
            ->send();
    }
}
