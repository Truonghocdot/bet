<?php

namespace App\Providers;

use App\Enum\User\RoleUser;
use App\Models\User;
use Illuminate\Foundation\Support\Providers\AuthServiceProvider as ServiceProvider;
use Illuminate\Support\Facades\Gate;

class AuthServiceProvider extends ServiceProvider
{
    public function boot(): void
    {
        Gate::before(function (User $user): ?bool {
            return $user->role === RoleUser::SUPER_ADMIN ? true : null;
        });

        $this->registerResourceAbilities();
        $this->registerCustomAbilities();
    }

    private function registerResourceAbilities(): void
    {
        $resourceAbilityMap = [
            'system.users.admins' => [RoleUser::SUPER_ADMIN],
            'system.users.clients' => [RoleUser::ADMIN, RoleUser::STAFF],
            'system.users.staffs' => [RoleUser::ADMIN],
            'system.users.agencies' => [RoleUser::ADMIN, RoleUser::STAFF],
            'finance.wallets' => [RoleUser::ADMIN, RoleUser::STAFF],
            'finance.wallet-ledger-entries' => [RoleUser::ADMIN, RoleUser::STAFF],
            'finance.transactions' => [RoleUser::ADMIN, RoleUser::STAFF],
            'finance.withdrawal-requests' => [RoleUser::ADMIN, RoleUser::STAFF],
            'finance.account-withdrawal-infos' => [RoleUser::ADMIN, RoleUser::STAFF],
            'system.notifications' => [RoleUser::ADMIN, RoleUser::STAFF],
            'system.banners' => [RoleUser::ADMIN, RoleUser::STAFF],
            'system.promotions' => [RoleUser::ADMIN, RoleUser::STAFF],
            'system.news-articles' => [RoleUser::ADMIN, RoleUser::STAFF],
            'payment.payment-receiving-accounts' => [RoleUser::ADMIN],
            'bet.game-periods' => [RoleUser::ADMIN, RoleUser::STAFF],
            'bet.bet-tickets' => [RoleUser::ADMIN, RoleUser::STAFF],
            'bet.bet-items' => [RoleUser::ADMIN, RoleUser::STAFF],
            'bet.bet-settlements' => [RoleUser::ADMIN, RoleUser::STAFF],
            'affiliate.affiliate-profiles' => [RoleUser::ADMIN, RoleUser::STAFF],
            'affiliate.affiliate-links' => [RoleUser::ADMIN, RoleUser::STAFF],
            'affiliate.affiliate-referrals' => [RoleUser::ADMIN, RoleUser::STAFF],
            'affiliate.affiliate-reward-settings' => [RoleUser::ADMIN],
            'affiliate.affiliate-reward-logs' => [RoleUser::ADMIN, RoleUser::STAFF],
        ];

        foreach ($resourceAbilityMap as $prefix => $allowedRoles) {
            foreach (['viewAny', 'view', 'create', 'update', 'delete', 'deleteAny', 'restore', 'restoreAny', 'forceDelete', 'forceDeleteAny'] as $ability) {
                Gate::define($prefix.'.'.$ability, fn (User $user): bool => in_array($user->role, $allowedRoles, true));
            }
        }
    }

    private function registerCustomAbilities(): void
    {
        $abilities = [
            'finance.withdrawal-requests.approve' => [RoleUser::ADMIN, RoleUser::STAFF],
            'finance.withdrawal-requests.reject' => [RoleUser::ADMIN, RoleUser::STAFF],
            'finance.withdrawal-requests.mark-paid' => [RoleUser::ADMIN],
            'bet.game-periods.settle' => [RoleUser::ADMIN, RoleUser::STAFF],
            'control_panel_access' => [RoleUser::ADMIN],
            'payment.payment-receiving-accounts.manage' => [RoleUser::ADMIN],
            'affiliate.affiliate-reward-settings.manage' => [RoleUser::ADMIN],
            'system.exchange-rate-settings.manage' => [RoleUser::ADMIN],
        ];

        foreach ($abilities as $ability => $allowedRoles) {
            Gate::define($ability, fn (User $user): bool => in_array($user->role, $allowedRoles, true));
        }
    }
}
