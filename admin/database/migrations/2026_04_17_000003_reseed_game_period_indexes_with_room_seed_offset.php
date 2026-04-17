<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Support\Facades\DB;

return new class extends Migration
{
    public function up(): void
    {
        DB::statement("
            with ranked_periods as (
                select
                    id,
                    (
                        to_char(draw_at at time zone 'Asia/Ho_Chi_Minh', 'YYYY')::bigint * 100000000
                    ) + (
                        10000000 + (
                            (('x' || substr(md5(lower(trim(room_code))), 1, 8))::bit(32)::bigint) % 50000000
                        )
                    ) + (
                        row_number() over (
                            partition by room_code, to_char(draw_at at time zone 'Asia/Ho_Chi_Minh', 'YYYY')
                            order by draw_at asc, id asc
                        ) - 1
                    ) as next_period_index
                from game_periods
            )
            update game_periods as gp
            set period_index = ranked_periods.next_period_index
            from ranked_periods
            where gp.id = ranked_periods.id
              and gp.period_index is distinct from ranked_periods.next_period_index
        ");
    }

    public function down(): void
    {
        // Irreversible data migration.
    }
};
