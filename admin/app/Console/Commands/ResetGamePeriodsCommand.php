<?php

namespace App\Console\Commands;

use Illuminate\Console\Command;
use Illuminate\Support\Facades\DB;
use Throwable;

class ResetGamePeriodsCommand extends Command
{
    protected $signature = 'bet:reset-game-periods
        {--force : Xoa du lieu ngay lap tuc, bo qua buoc xac nhan}
        {--dry-run : Chi hien thi so dong se bi xoa}';

    protected $description = 'Xoa toan bo game periods va du lieu bet lien quan, reset id game_periods ve 0';

    /**
     * @var list<string>
     */
    private array $tables = [
        'bet_settlements',
        'bet_items',
        'bet_tickets',
        'game_periods',
    ];

    public function handle(): int
    {
        $driver = DB::connection()->getDriverName();
        $counts = $this->tableCounts();

        $this->warn('Command nay se xoa toan bo du lieu lien quan den ky quay va lenh bet.');
        $this->table(
            ['Table', 'Rows'],
            array_map(
                static fn (string $table, int $count): array => [$table, (string) $count],
                array_keys($counts),
                array_values($counts),
            ),
        );
        $this->line(sprintf('Database driver hien tai: %s', $driver));

        if ($this->option('dry-run')) {
            $this->info('Dry-run: khong co du lieu nao bi xoa.');

            return self::SUCCESS;
        }

        if (! $this->option('force') && ! $this->confirm('Ban chac chan muon xoa tat ca du lieu ben tren?')) {
            $this->warn('Da huy command.');

            return self::SUCCESS;
        }

        try {
            match ($driver) {
                'pgsql' => $this->resetForPostgres(),
                'mysql', 'mariadb' => $this->resetForMySql(),
                'sqlite' => $this->resetForSqlite(),
                default => $this->resetByDelete(),
            };
        } catch (Throwable $exception) {
            $this->error($exception->getMessage());

            return self::FAILURE;
        }

        $this->info('Da xoa du lieu game periods, cac bang phu thuoc va reset sequence cua game_periods.id.');
        $this->line('Luu y: period_index khong phai auto increment cua database, no se duoc app tu dong gan lai tu 0 khi tao ky moi.');

        return self::SUCCESS;
    }

    /**
     * @return array<string, int>
     */
    private function tableCounts(): array
    {
        $counts = [];

        foreach ($this->tables as $table) {
            $counts[$table] = DB::table($table)->count();
        }

        return $counts;
    }

    private function resetForPostgres(): void
    {
        DB::statement(sprintf('TRUNCATE TABLE %s RESTART IDENTITY', implode(', ', $this->tables)));

        $sequence = DB::scalar("select pg_get_serial_sequence('game_periods', 'id')");

        if (! is_string($sequence) || $sequence === '') {
            return;
        }

        DB::statement(sprintf('ALTER SEQUENCE %s MINVALUE 0 RESTART WITH 0', $sequence));
    }

    private function resetForMySql(): void
    {
        DB::statement('SET FOREIGN_KEY_CHECKS=0');

        try {
            foreach ($this->tables as $table) {
                DB::table($table)->truncate();
            }
        } finally {
            DB::statement('SET FOREIGN_KEY_CHECKS=1');
        }
    }

    private function resetForSqlite(): void
    {
        DB::statement('PRAGMA foreign_keys = OFF');

        try {
            foreach ($this->tables as $table) {
                DB::table($table)->delete();
            }

            $hasSqliteSequence = DB::table('sqlite_master')
                ->where('type', 'table')
                ->where('name', 'sqlite_sequence')
                ->exists();

            if ($hasSqliteSequence) {
                DB::table('sqlite_sequence')
                    ->whereIn('name', $this->tables)
                    ->delete();
            }
        } finally {
            DB::statement('PRAGMA foreign_keys = ON');
        }
    }

    private function resetByDelete(): void
    {
        foreach ($this->tables as $table) {
            DB::table($table)->delete();
        }
    }
}
