<?php

namespace App\Filament\Pages;

use App\Filament\Widgets\OperationsStatsWidget;
use Filament\Pages\Dashboard as PagesDashboard;

class Dashboard extends PagesDashboard
{
    protected static ?string $navigationLabel = 'Bảng điều khiển';

    protected static ?string $title = 'Bảng điều khiển';

    /**
     * @return array<class-string<\Filament\Widgets\Widget>>
     */
    public function getWidgets(): array
    {
        return [
            OperationsStatsWidget::class,
        ];
    }

    public function getColumns(): int|array
    {
        return 1;
    }
}
