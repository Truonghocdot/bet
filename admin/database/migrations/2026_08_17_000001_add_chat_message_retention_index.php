<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Support\Facades\DB;

return new class extends Migration
{
    public function up(): void
    {
        DB::statement('create index if not exists chat_messages_created_at_index on chat_messages (created_at)');
    }

    public function down(): void
    {
        DB::statement('drop index if exists chat_messages_created_at_index');
    }
};
