<?php

namespace App\Console\Commands;

use App\Enum\Bet\BetStatus;
use App\Models\Bet\BetTicket;
use App\Models\Bet\GameRoundHistory;
use Illuminate\Console\Command;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Log;
use Throwable;

class CleanupBetDataCommand extends Command
{
    protected $signature = 'bet:cleanup
        {--history-days=30 : So ngay giu lai lich su ket qua}
        {--bet-days=30 : So ngay giu lai lenh bet da ket thuc}
        {--dry-run : Chi thong ke, khong xoa du lieu}';

    protected $description = 'Don dep ket qua va lenh bet da het han';

    public function handle(): int
    {
        $historyDays = max(1, (int) $this->option('history-days'));
        $betDays = max(1, (int) $this->option('bet-days'));
        $dryRun = (bool) $this->option('dry-run');

        $historyCutoff = now()->subDays($historyDays);
        $betCutoff = now()->subDays($betDays);
        $terminalStatusValues = $this->terminalStatusValues();

        $this->info(sprintf(
            'Bat dau don dep bet/result. history_days=%d bet_days=%d dry_run=%s cutoff_history=%s cutoff_bet=%s',
            $historyDays,
            $betDays,
            $dryRun ? 'true' : 'false',
            $historyCutoff->toDateTimeString(),
            $betCutoff->toDateTimeString(),
        ));

        Log::info('bet.cleanup.started', [
            'history_days' => $historyDays,
            'bet_days' => $betDays,
            'dry_run' => $dryRun,
            'history_cutoff' => $historyCutoff->toDateTimeString(),
            'bet_cutoff' => $betCutoff->toDateTimeString(),
        ]);

        try {
            $historyCount = GameRoundHistory::query()
                ->where('draw_at', '<', $historyCutoff)
                ->count();

            $ticketQuery = BetTicket::query()
                ->whereIn('status', $terminalStatusValues)
                ->whereRaw('COALESCE(settled_at, updated_at, created_at) < ?', [$betCutoff]);

            $ticketCount = (clone $ticketQuery)->count();

            $this->line(sprintf(
                'Se xoa: histories=%d tickets=%d settlements=tu tickets, items=tu tickets',
                $historyCount,
                $ticketCount,
            ));

            Log::info('bet.cleanup.candidates', [
                'history_count' => $historyCount,
                'ticket_count' => $ticketCount,
            ]);

            if ($dryRun) {
                $this->warn('Dry-run: khong thuc hien xoa du lieu.');

                return self::SUCCESS;
            }

            $deletedHistoryCount = GameRoundHistory::query()
                ->where('draw_at', '<', $historyCutoff)
                ->delete();

            $deletedTickets = 0;
            $deletedItems = 0;
            $deletedSettlements = 0;
            $chunks = 0;

            $ticketQuery
                ->select(['id'])
                ->orderBy('id')
                ->chunkById(500, function ($tickets) use (
                    &$deletedTickets,
                    &$deletedItems,
                    &$deletedSettlements,
                    &$chunks
                ): void {
                    $ticketIds = $tickets->pluck('id')->all();
                    if (empty($ticketIds)) {
                        return;
                    }

                    $chunks++;
                    Log::info('bet.cleanup.chunk.start', [
                        'chunk' => $chunks,
                        'ticket_ids' => $ticketIds,
                    ]);

                    DB::transaction(function () use ($ticketIds, &$deletedTickets, &$deletedItems, &$deletedSettlements): void {
                        $deletedSettlements += DB::table('bet_settlements')
                            ->whereIn('ticket_id', $ticketIds)
                            ->delete();

                        $deletedItems += DB::table('bet_items')
                            ->whereIn('ticket_id', $ticketIds)
                            ->delete();

                        $deletedTickets += DB::table('bet_tickets')
                            ->whereIn('id', $ticketIds)
                            ->delete();
                    });

                    Log::info('bet.cleanup.chunk.done', [
                        'chunk' => $chunks,
                        'deleted_tickets_total' => $deletedTickets,
                        'deleted_items_total' => $deletedItems,
                        'deleted_settlements_total' => $deletedSettlements,
                    ]);
                });

            $this->info(sprintf(
                'Hoan tat don dep. histories=%d tickets=%d items=%d settlements=%d',
                $deletedHistoryCount,
                $deletedTickets,
                $deletedItems,
                $deletedSettlements,
            ));

            Log::info('bet.cleanup.completed', [
                'deleted_histories' => $deletedHistoryCount,
                'deleted_tickets' => $deletedTickets,
                'deleted_items' => $deletedItems,
                'deleted_settlements' => $deletedSettlements,
            ]);

            return self::SUCCESS;
        } catch (Throwable $exception) {
            Log::error('bet.cleanup.failed', [
                'message' => $exception->getMessage(),
                'exception' => $exception::class,
            ]);

            $this->error($exception->getMessage());

            return self::FAILURE;
        }
    }

    /**
     * @return array<int, int>
     */
    private function terminalStatusValues(): array
    {
        return array_values(array_map(
            static fn (BetStatus $status): int => $status->value,
            array_filter(
                BetStatus::cases(),
                static fn (BetStatus $status): bool => $status !== BetStatus::PENDING,
            ),
        ));
    }
}
