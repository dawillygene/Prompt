<?php

use Illuminate\Database\Migrations\Migration;
use Illuminate\Database\Schema\Blueprint;
use Illuminate\Support\Facades\Schema;

return new class extends Migration
{
    public function up(): void
    {
        Schema::create('categories', function (Blueprint $table): void {
            $table->id();
            $table->foreignId('user_id')->constrained()->cascadeOnDelete();
            $table->string('name', 100);
            $table->string('slug');
            $table->text('description')->nullable();
            $table->timestamps();
            $table->unique(['user_id', 'slug']);
        });

        Schema::create('tags', function (Blueprint $table): void {
            $table->id();
            $table->foreignId('user_id')->constrained()->cascadeOnDelete();
            $table->string('name', 100);
            $table->string('slug');
            $table->timestamps();
            $table->unique(['user_id', 'slug']);
        });

        Schema::create('prompts', function (Blueprint $table): void {
            $table->id();
            $table->foreignId('user_id')->constrained()->cascadeOnDelete();
            $table->string('title');
            $table->string('slug');
            $table->text('summary')->nullable();
            $table->longText('content');
            $table->foreignId('category_id')->nullable()->constrained()->nullOnDelete();
            $table->string('visibility', 20)->default('private');
            $table->boolean('is_favorite')->default(false);
            $table->boolean('is_archived')->default(false);
            $table->timestamps();
            $table->softDeletes();
            $table->unique(['user_id', 'slug'])->whereNull('deleted_at');
        });

        Schema::create('prompt_tag', function (Blueprint $table): void {
            $table->id();
            $table->foreignId('prompt_id')->constrained()->cascadeOnDelete();
            $table->foreignId('tag_id')->constrained()->cascadeOnDelete();
            $table->unique(['prompt_id', 'tag_id']);
        });

        Schema::create('prompt_versions', function (Blueprint $table): void {
            $table->id();
            $table->foreignId('prompt_id')->constrained()->cascadeOnDelete();
            $table->longText('content');
            $table->unsignedInteger('version_number');
            $table->timestamp('created_at')->useCurrent();
            $table->unique(['prompt_id', 'version_number']);
        });

        Schema::create('audit_logs', function (Blueprint $table): void {
            $table->id();
            $table->foreignId('user_id')->constrained()->cascadeOnDelete();
            $table->string('action');
            $table->string('entity_type');
            $table->unsignedBigInteger('entity_id');
            $table->json('metadata')->nullable();
            $table->timestamp('created_at')->useCurrent();
            $table->index(['user_id', 'action']);
            $table->index(['entity_type', 'entity_id']);
        });
    }

    public function down(): void
    {
        Schema::dropIfExists('audit_logs');
        Schema::dropIfExists('prompt_versions');
        Schema::dropIfExists('prompt_tag');
        Schema::dropIfExists('prompts');
        Schema::dropIfExists('tags');
        Schema::dropIfExists('categories');
    }
};
