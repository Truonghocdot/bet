<?php

namespace App\Filament\Pages\Agency;

use App\Enum\Transaction\TransactionStatus;
use App\Enum\Transaction\TypeTransaction;
use App\Enum\Transaction\WithdrawalStatus;
use App\Enum\User\RoleUser;
use App\Models\Transaction\Transaction;
use App\Models\Transaction\WithdrawalRequest;
use App\Models\User;
use App\Support\Filament\EnumPresenter;
use BackedEnum;
use Filament\Forms\Components\Select;
use Filament\Pages\Concerns\HasMaxWidth;
use Filament\Pages\Concerns\HasTopbar;
use Filament\Pages\Page;
use Filament\Schemas\Components\Section;
use Filament\Schemas\Schema;
use Filament\Support\Icons\Heroicon;
use Illuminate\Contracts\Pagination\LengthAwarePaginator;
use Illuminate\Contracts\Support\Htmlable;
use Illuminate\Database\Eloquent\Builder;
use Illuminate\Support\Facades\Gate;
use UnitEnum;

class AgencyBackofficePage extends Page
{
    use HasMaxWidth;
    use HasTopbar;

    protected string $view = 'filament.pages.agency.backoffice-page';

    protected static bool $isDiscovered = true;

    protected static ?string $slug = 'agency/backoffice';

    protected static UnitEnum|string|null $navigationGroup = 'Tài chính';

    protected static ?string $navigationLabel = 'Thống kê Đại lý';

    protected static ?string $title = 'Thống kê Đại lý';

    protected static string|BackedEnum|null $navigationIcon = Heroicon::OutlinedBuildingOffice2;

    protected static ?int $navigationSort = 50;

    public ?array $data = [];

    public function mount(): void
    {
        $defaultAgencyId = User::query()
            ->where('role', RoleUser::AGENCY->value)
            ->orderBy('name')
            ->value('id');

        $this->data = [
            'agency_id' => $defaultAgencyId,
        ];
    }

    public static function canAccess(): bool
    {
        return Gate::allows('agency.backoffice.view');
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
        return 'Admin/staff chọn đại lý để xem người chơi, giao dịch và yêu cầu rút thuộc tuyến đó.';
    }

    public function defaultForm(Schema $schema): Schema
    {
        return $schema->statePath('data');
    }

    public function form(Schema $schema): Schema
    {
        return $schema->components([
            Section::make('Bộ lọc đại lý')
                ->schema([
                    Select::make('agency_id')
                        ->label('Chọn đại lý')
                        ->options($this->agencyOptions())
                        ->searchable()
                        ->preload()
                        ->live()
                        ->required(),
                ]),
        ]);
    }

    public function getViewData(): array
    {
        $agency = $this->selectedAgency();

        return [
            'agency' => $agency,
            'summary' => [
                'players' => $agency ? $this->playersBaseQuery($agency->id)->count() : 0,
                'transactions' => $agency ? $this->transactionsBaseQuery($agency->id)->count() : 0,
                'withdrawals' => $agency ? $this->withdrawalsBaseQuery($agency->id)->count() : 0,
            ],
            'players' => $agency ? $this->players($agency->id) : null,
            'transactions' => $agency ? $this->transactions($agency->id) : null,
            'withdrawals' => $agency ? $this->withdrawals($agency->id) : null,
        ];
    }

    /**
     * @return array<int, string>
     */
    protected function agencyOptions(): array
    {
        return User::query()
            ->where('role', RoleUser::AGENCY->value)
            ->orderBy('name')
            ->get(['id', 'name'])
            ->mapWithKeys(fn (User $agency): array => [
                $agency->id => sprintf('%s (#%d)', $agency->name ?: 'Đại lý', $agency->id),
            ])
            ->all();
    }

    protected function selectedAgency(): ?User
    {
        $agencyId = (int) ($this->data['agency_id'] ?? 0);
        if ($agencyId <= 0) {
            return null;
        }

        return User::query()
            ->with('affiliateProfile')
            ->whereKey($agencyId)
            ->where('role', RoleUser::AGENCY->value)
            ->first();
    }

    protected function players(int $agencyId): LengthAwarePaginator
    {
        return $this->playersBaseQuery($agencyId)
            ->with(['affiliateProfile'])
            ->select(['id', 'name', 'phone', 'status', 'created_at'])
            ->orderByDesc('id')
            ->paginate(10, ['*'], 'players_page');
    }

    protected function transactions(int $agencyId): LengthAwarePaginator
    {
        return $this->transactionsBaseQuery($agencyId)
            ->with(['user'])
            ->select([
                'id',
                'user_id',
                'client_ref',
                'type',
                'unit',
                'amount',
                'status',
                'provider',
                'created_at',
            ])
            ->orderByDesc('id')
            ->paginate(10, ['*'], 'transactions_page');
    }

    protected function withdrawals(int $agencyId): LengthAwarePaginator
    {
        return $this->withdrawalsBaseQuery($agencyId)
            ->with(['user'])
            ->select([
                'id',
                'user_id',
                'unit',
                'amount',
                'fee',
                'net_amount',
                'status',
                'created_at',
            ])
            ->orderByDesc('id')
            ->paginate(10, ['*'], 'withdrawals_page');
    }

    protected function playersBaseQuery(int $agencyId): Builder
    {
        return User::query()->whereHas('referredByReferral', function (Builder $query) use ($agencyId): void {
            $query->where('referrer_user_id', $agencyId);
        });
    }

    protected function transactionsBaseQuery(int $agencyId): Builder
    {
        return Transaction::query()->whereHas('user.referredByReferral', function (Builder $query) use ($agencyId): void {
            $query->where('referrer_user_id', $agencyId);
        });
    }

    protected function withdrawalsBaseQuery(int $agencyId): Builder
    {
        return WithdrawalRequest::query()->whereHas('user.referredByReferral', function (Builder $query) use ($agencyId): void {
            $query->where('referrer_user_id', $agencyId);
        });
    }

    public function transactionTypeLabel(int $value): string
    {
        return EnumPresenter::label(TypeTransaction::class, $value);
    }

    public function transactionStatusLabel(int $value): string
    {
        return EnumPresenter::label(TransactionStatus::class, $value);
    }

    public function withdrawalStatusLabel(int $value): string
    {
        return EnumPresenter::label(WithdrawalStatus::class, $value);
    }
}
