<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('sync_queues', function (Blueprint $table) {
            $table->id();
            $table->unsignedBigInteger('user_id');
            $table->enum('action', ['create', 'update', 'delete']); // Type of operation
            $table->string('entity_type'); // 'prompt', 'category', 'tag'
            $table->unsignedBigInteger('entity_id'); // ID of the entity
            $table->json('payload')->nullable(); // Serialized data for create/update
            $table->enum('status', ['pending', 'synced', 'conflict'])->default('pending');
            $table->unsignedBigInteger('conflict_with_id')->nullable(); // Server entity ID if conflict
            $table->enum('conflict_type', ['newer_on_server', 'deleted_on_server', 'both_modified'])->nullable();
            $table->timestamps();

            $table->foreign('user_id')->references('id')->on('users')->cascadeOnDelete();
            $table->index(['user_id', 'status']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('sync_queues');
    }
};
