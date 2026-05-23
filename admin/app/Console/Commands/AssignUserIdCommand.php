<?php

namespace App\Console\Commands;

use App\Models\User;
use Illuminate\Console\Command;
use Illuminate\Support\Carbon;
use Illuminate\Support\Facades\DB;
use InvalidArgumentException;
use Throwable;

class AssignUserIdCommand extends Command
{
    protected $signature = 'users:assign-id
        {user_id? : User ID hiện tại cần đổi}
        {--new-id= : Gán cố định một user ID 6 số}
        {--force : Xử lý cả user đã có ID 6 số}
        {--dry-run : Chỉ in ra kế hoạch, không ghi dữ liệu}';

    protected $description = 'Đổi user.id sang mã 6 số không trùng nhau và cập nhật toàn bộ khóa ngoại liên quan.';

    /**
     * @var array<int, array{table:string,column:string}>
     */
    private array $references = [
        ['table' => 'account_withdrawal_infos', 'column' => 'user_id'],
        ['table' => 'affiliate_profiles', 'column' => 'user_id'],
        ['table' => 'affiliate_referrals', 'column' => 'referrer_user_id'],
        ['table' => 'affiliate_referrals', 'column' => 'referred_user_id'],
        ['table' => 'affiliate_reward_logs', 'column' => 'referrer_user_id'],
        ['table' => 'auth_otp_requests', 'column' => 'user_id'],
        ['table' => 'auth_refresh_tokens', 'column' => 'user_id'],
        ['table' => 'banners', 'column' => 'created_by'],
        ['table' => 'bet_settlements', 'column' => 'settled_by'],
        ['table' => 'bet_tickets', 'column' => 'user_id'],
        ['table' => 'exchange_rate_settings', 'column' => 'updated_by'],
        ['table' => 'news_articles', 'column' => 'created_by'],
        ['table' => 'notification_reads', 'column' => 'user_id'],
        ['table' => 'notification_targets', 'column' => 'user_id'],
        ['table' => 'notifications', 'column' => 'created_by'],
        ['table' => 'transactions', 'column' => 'approved_by'],
        ['table' => 'transactions', 'column' => 'user_id'],
        ['table' => 'wallet_ledger_entries', 'column' => 'user_id'],
        ['table' => 'wallets', 'column' => 'user_id'],
        ['table' => 'withdrawal_requests', 'column' => 'paid_by'],
        ['table' => 'withdrawal_requests', 'column' => 'reviewed_by'],
        ['table' => 'withdrawal_requests', 'column' => 'user_id'],
    ];

    public function handle(): int
    {
        try {
            $targets = $this->resolveTargets();
        } catch (InvalidArgumentException $exception) {
            $this->error($exception->getMessage());

            return self::INVALID;
        }

        if ($targets === []) {
            $this->warn('Không có user nào cần xử lý.');

            return self::SUCCESS;
        }

        foreach ($targets as $target) {
            $oldId = (int) $target->id;
            $newId = (int) $target->new_id;

            $this->line(sprintf('User %d -> %d (%s)', $oldId, $newId, $target->name ?? 'N/A'));

            if ($this->option('dry-run')) {
                continue;
            }

            try {
                $this->reassignUserId($oldId, $newId);
            } catch (Throwable $exception) {
                $this->error(sprintf('Đổi ID user %d thất bại: %s', $oldId, $exception->getMessage()));

                return self::FAILURE;
            }
        }

        if ($this->option('dry-run')) {
            $this->info('Dry-run hoàn tất, chưa ghi thay đổi.');
            return self::SUCCESS;
        }

        $this->warn('Lưu ý: access token cũ của user đã đổi ID sẽ không còn hợp lệ và cần đăng nhập lại.');
        $this->info('Đã cập nhật user ID thành công.');

        return self::SUCCESS;
    }

    /**
     * @return list<object{id:int,name:string,new_id:int}>
     */
    private function resolveTargets(): array
    {
        $inputId = $this->argument('user_id');
        $manualNewId = $this->option('new-id');
        $force = (bool) $this->option('force');

        if ($manualNewId !== null && $inputId === null) {
            throw new InvalidArgumentException('Phải truyền `user_id` khi dùng `--new-id`.');
        }

        if ($inputId !== null) {
            $user = DB::table('users')
                ->select(['id', 'name'])
                ->where('id', (int) $inputId)
                ->first();

            if ($user === null) {
                throw new InvalidArgumentException('Không tìm thấy user cần đổi ID.');
            }

            $oldId = (int) $user->id;
            if (! $force && $manualNewId === null && $this->isSixDigitId($oldId)) {
                return [];
            }

            $newId = $manualNewId !== null
                ? $this->parseSixDigitId((string) $manualNewId, $oldId)
                : User::generateUniqueSixDigitId($oldId);

            return [(object) [
                'id' => $oldId,
                'name' => (string) $user->name,
                'new_id' => $newId,
            ]];
        }

        $users = DB::table('users')
            ->select(['id', 'name'])
            ->when(! $force, fn ($query) => $query->where(function ($inner): void {
                $inner->where('id', '<', 100000)->orWhere('id', '>', 999999);
            }))
            ->orderBy('id')
            ->get();

        $reservedIds = [];
        $targets = [];

        foreach ($users as $user) {
            $newId = User::generateUniqueSixDigitId((int) $user->id);
            while (isset($reservedIds[$newId])) {
                $newId = User::generateUniqueSixDigitId((int) $user->id);
            }

            $reservedIds[$newId] = true;
            $targets[] = (object) [
                'id' => (int) $user->id,
                'name' => (string) $user->name,
                'new_id' => $newId,
            ];
        }

        return $targets;
    }

    private function parseSixDigitId(string $value, int $oldId): int
    {
        if (! preg_match('/^\d{6}$/', $value)) {
            throw new InvalidArgumentException('`--new-id` phải là số gồm đúng 6 chữ số.');
        }

        $newId = (int) $value;
        if ($newId === $oldId) {
            throw new InvalidArgumentException('`--new-id` phải khác ID hiện tại.');
        }

        $exists = User::query()->whereKey($newId)->exists();
        if ($exists) {
            throw new InvalidArgumentException('`--new-id` đã tồn tại.');
        }

        return $newId;
    }

    private function isSixDigitId(int $value): bool
    {
        return $value >= 100000 && $value <= 999999;
    }

    private function reassignUserId(int $oldId, int $newId): void
    {
        DB::transaction(function () use ($oldId, $newId): void {
            $user = DB::table('users')->where('id', $oldId)->lockForUpdate()->first();
            if ($user === null) {
                throw new InvalidArgumentException(sprintf('Không tìm thấy user %d.', $oldId));
            }

            if (DB::table('users')->where('id', $newId)->exists()) {
                throw new InvalidArgumentException(sprintf('User ID %d đã tồn tại.', $newId));
            }

            $now = Carbon::now();
            $tempEmail = $user->email !== null ? sprintf('migrating+%d-%d@local.invalid', $oldId, $newId) : null;

            DB::table('users')
                ->where('id', $oldId)
                ->update([
                    'email' => $tempEmail,
                    'phone' => null,
                    'updated_at' => $now,
                ]);

            DB::table('users')->insert([
                'id' => $newId,
                'name' => $user->name,
                'email' => $user->email,
                'phone' => $user->phone,
                'email_verified_at' => $user->email_verified_at,
                'phone_verified_at' => $user->phone_verified_at,
                'password' => $user->password,
                'remember_token' => $user->remember_token,
                'role' => $user->role,
                'status' => $user->status,
                'last_login_at' => $user->last_login_at,
                'created_at' => $user->created_at,
                'updated_at' => $user->updated_at,
                'deleted_at' => $user->deleted_at,
            ]);

            foreach ($this->references as $reference) {
                DB::table($reference['table'])
                    ->where($reference['column'], $oldId)
                    ->update([$reference['column'] => $newId]);
            }

            DB::table('users')->where('id', $oldId)->delete();
        });
    }
}
