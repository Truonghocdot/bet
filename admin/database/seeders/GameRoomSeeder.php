<?php

namespace Database\Seeders;

use App\Enum\Bet\GameType;
use App\Enum\Bet\RoomStatus;
use App\Models\Bet\GameRoom;
use Illuminate\Database\Seeder;

class GameRoomSeeder extends Seeder
{
    public function run(): void
    {
        $rooms = [
            ['code' => 'wingo_30s', 'game_type' => GameType::WINGO, 'duration_seconds' => 30, 'sort_order' => 1],
            ['code' => 'wingo_1m', 'game_type' => GameType::WINGO, 'duration_seconds' => 60, 'sort_order' => 2],
            ['code' => 'wingo_3m', 'game_type' => GameType::WINGO, 'duration_seconds' => 180, 'sort_order' => 3],
            ['code' => 'wingo_5m', 'game_type' => GameType::WINGO, 'duration_seconds' => 300, 'sort_order' => 4],
            ['code' => 'k3_1m', 'game_type' => GameType::K3, 'duration_seconds' => 60, 'sort_order' => 5],
            ['code' => 'k3_3m', 'game_type' => GameType::K3, 'duration_seconds' => 180, 'sort_order' => 6],
            ['code' => 'k3_5m', 'game_type' => GameType::K3, 'duration_seconds' => 300, 'sort_order' => 7],
            ['code' => 'k3_10m', 'game_type' => GameType::K3, 'duration_seconds' => 600, 'sort_order' => 8],
            ['code' => 'lottery_1m', 'game_type' => GameType::LOTTERY, 'duration_seconds' => 60, 'sort_order' => 9],
            ['code' => 'lottery_3m', 'game_type' => GameType::LOTTERY, 'duration_seconds' => 180, 'sort_order' => 10],
            ['code' => 'lottery_5m', 'game_type' => GameType::LOTTERY, 'duration_seconds' => 300, 'sort_order' => 11],
            ['code' => 'lottery_10m', 'game_type' => GameType::LOTTERY, 'duration_seconds' => 600, 'sort_order' => 12],
        ];

        foreach ($rooms as $room) {
            GameRoom::query()->updateOrCreate(
                ['code' => $room['code']],
                [
                    'game_type' => $room['game_type'],
                    'duration_seconds' => $room['duration_seconds'],
                    'bet_cutoff_seconds' => 5,
                    'status' => RoomStatus::ACTIVE,
                    'sort_order' => $room['sort_order'],
                ],
            );
        }
    }
}
