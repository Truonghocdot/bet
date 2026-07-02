<?php

namespace App\Console\Commands;

use App\Enum\User\RoleUser;
use App\Enum\User\UserStatus;
use App\Models\User;
use Illuminate\Console\Command;
use Illuminate\Database\Eloquent\Builder;
use Illuminate\Support\Facades\Cache;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Str;

class SyncTcgUsersCommand extends Command
{
    protected $signature = 'users:sync-tcg-register
        {--user-id=* : Chỉ đồng bộ một hoặc nhiều user id cụ thể}
        {--roles=client,agency : Danh sách role cần đồng bộ, ví dụ client,agency}
        {--chunk=200 : Số lượng user xử lý mỗi lần}
        {--include-inactive : Bao gồm cả user không ở trạng thái active}
        {--dry-run : Chỉ in ra username/password tcg, không gọi API}
        {--stop-on-error : Dừng ngay khi gặp lỗi}';

    protected $description = 'Đồng bộ toàn bộ user hiện có sang TCG qua gate internal API.';

    public function handle(): int
    {
        $lock = Cache::store('redis')->lock('lock:users:sync-tcg-register', 900);

        if (! $lock->get()) {
            $this->warn('Đang có một tiến trình đồng bộ TCG khác chạy.');

            return self::SUCCESS;
        }

        try {
            $baseUrl = rtrim((string) config('services.tcg.gate_base_url'), '/');
            $token = trim((string) config('services.tcg.gate_internal_token'));
            $timeout = (int) config('services.tcg.timeout', 15);
            $currency = trim((string) config('services.tcg.default_currency', 'VND'));

            if ($baseUrl === '') {
                $this->error('Thiếu cấu hình services.tcg.gate_base_url');

                return self::FAILURE;
            }

            if ($token === '') {
                $this->error('Thiếu cấu hình services.tcg.gate_internal_token');

                return self::FAILURE;
            }

            $query = $this->buildUserQuery();
            $total = (clone $query)->count();

            if ($total === 0) {
                $this->warn('Không có user nào phù hợp để đồng bộ.');

                return self::SUCCESS;
            }

            $chunkSize = max(1, (int) $this->option('chunk'));
            $processed = 0;
            $success = 0;
            $failed = 0;
            $skipped = 0;

            $this->info(sprintf('Bắt đầu đồng bộ %d user sang TCG.', $total));

            $query
                ->orderBy('id')
                ->chunkById($chunkSize, function ($users) use ($baseUrl, $token, $timeout, $currency, &$processed, &$success, &$failed, &$skipped): bool {
                    foreach ($users as $user) {
                        $processed++;

                        $username = $this->buildTcgUsername($user);
                        $password = $this->buildTcgPassword($username);

                        if ($username === '') {
                            $skipped++;
                            $this->warn(sprintf('[SKIP] user_id=%d reason=empty_username', $user->id));
                            continue;
                        }

                        if ($this->option('dry-run')) {
                            $this->line(sprintf(
                                '[DRY-RUN] user_id=%d role=%s username=%s password=%s',
                                $user->id,
                                $user->role->name,
                                $username,
                                $password,
                            ));
                            $success++;
                            continue;
                        }

                        $response = Http::timeout($timeout)
                            ->acceptJson()
                            ->withHeaders([
                                'X-Internal-Token' => $token,
                            ])
                            ->post($baseUrl.'/internal/v1/tcg/players/register', [
                                'username' => $username,
                                'password' => $password,
                                'currency' => $currency,
                            ]);

                        if ($response->successful()) {
                            $success++;
                            $this->info(sprintf('[OK] user_id=%d username=%s', $user->id, $username));
                            continue;
                        }

                        $failed++;
                        $message = trim((string) ($response->json('message') ?? $response->body() ?? 'Unknown error'));
                        $this->error(sprintf('[FAIL] user_id=%d username=%s status=%d message=%s', $user->id, $username, $response->status(), $message));

                        if ($this->option('stop-on-error')) {
                            return false;
                        }
                    }

                    return true;
                });

            $this->newLine();
            $this->table(
                ['Tổng', 'Thành công', 'Thất bại', 'Bỏ qua'],
                [[(string) $processed, (string) $success, (string) $failed, (string) $skipped]],
            );

            return $failed > 0 && $this->option('stop-on-error')
                ? self::FAILURE
                : self::SUCCESS;
        } finally {
            optional($lock)->release();
        }
    }

    private function buildUserQuery(): Builder
    {
        $roles = $this->resolveRoles((string) $this->option('roles'));
        $userIds = collect((array) $this->option('user-id'))
            ->map(static fn ($value): int => (int) $value)
            ->filter(static fn (int $value): bool => $value > 0)
            ->values();

        return User::query()
            ->when($userIds->isNotEmpty(), fn (Builder $query) => $query->whereIn('id', $userIds->all()))
            ->whereIn('role', array_map(static fn (RoleUser $role): int => $role->value, $roles))
            ->when(! $this->option('include-inactive'), fn (Builder $query) => $query->where('status', UserStatus::ACTIVE->value));
    }

    /**
     * @return list<RoleUser>
     */
    private function resolveRoles(string $rawRoles): array
    {
        $aliases = [
            'client' => RoleUser::CLIENT,
            'agency' => RoleUser::AGENCY,
        ];

        $roles = collect(explode(',', $rawRoles))
            ->map(static fn (string $value): string => strtolower(trim($value)))
            ->filter()
            ->map(static fn (string $value) => $aliases[$value] ?? null)
            ->filter()
            ->unique(static fn (RoleUser $role): int => $role->value)
            ->values()
            ->all();

        return $roles !== [] ? $roles : [RoleUser::CLIENT, RoleUser::AGENCY];
    }

    private function buildTcgUsername(User $user): string
    {
        $base = $this->sanitizeTcgUsernameBase((string) $user->name);
        if ($base === '') {
            $base = 'user';
        }

        $phoneSeed = preg_replace('/\D+/', '', (string) ($user->phone ?? '')) ?: '';
        $hashSuffix = substr(sha1($phoneSeed.':'.$base), 0, 8);
        $username = $base.$hashSuffix;

        if (strlen($username) > 14) {
            $baseLimit = max(1, 14 - strlen($hashSuffix));
            $base = substr($base, 0, $baseLimit);
            $username = $base.$hashSuffix;
        }

        while (strlen($username) < 4) {
            $username .= '9';
        }

        return $username;
    }

    private function sanitizeTcgUsernameBase(string $value): string
    {
        $ascii = Str::of($value)
            ->ascii()
            ->lower()
            ->replaceMatches('/[^a-z0-9]+/', '')
            ->value();

        return substr($ascii, 0, 6);
    }

    private function buildTcgPassword(string $username): string
    {
        $password = substr('Tcg'.$username, 0, 12);

        while (strlen($password) < 6) {
            $password .= '9';
        }

        return $password;
    }
}
