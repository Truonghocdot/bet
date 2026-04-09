<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('withdrawal_requests', function (Blueprint $table) {
            $table->foreignId('paid_by')->nullable()->after('reviewed_at')->constrained('users')->nullOnDelete();
            $table->timestamp('paid_at')->nullable()->after('paid_by');
            $table->string('transfer_reference', 255)->nullable()->after('paid_at');
            $table->string('transfer_proof_path', 255)->nullable()->after('transfer_reference');
            $table->text('admin_note')->nullable()->after('transfer_proof_path');
        });
    }

    public function down(): void
    {
        Schema::table('withdrawal_requests', function (Blueprint $table) {
            $table->dropConstrainedForeignId('paid_by');
            $table->dropColumn([
                'paid_at',
                'transfer_reference',
                'transfer_proof_path',
                'admin_note',
            ]);
        });
    }
};
