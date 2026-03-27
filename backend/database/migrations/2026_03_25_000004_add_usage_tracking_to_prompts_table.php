<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::table('prompts', function (Blueprint $table) {
            $table->unsignedBigInteger('usage_count')->default(0)->after('is_archived');
            $table->timestamp('last_used_at')->nullable()->after('usage_count');
        });
    }

    public function down(): void
    {
        Schema::table('prompts', function (Blueprint $table) {
            $table->dropColumn('usage_count');
            $table->dropColumn('last_used_at');
        });
    }
};
