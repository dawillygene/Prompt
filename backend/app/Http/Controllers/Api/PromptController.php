<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use App\Http\Requests\StorePromptRequest;
use App\Http\Requests\UpdatePromptRequest;
use App\Models\AuditLog;
use App\Models\Prompt;
use App\Models\PromptVersion;
use App\Models\User;
use Illuminate\Database\Eloquent\Builder;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Str;

class PromptController extends Controller
{
    public function index(Request $request): JsonResponse
    {
        /** @var User $user */
        $user = $request->user();

        $query = Prompt::query()
            ->with(['category', 'tags'])
            ->where('user_id', $user->id);

        if ($search = $request->string('search')->toString()) {
            $query->where(function (Builder $builder) use ($search): void {
                $builder
                    ->where('title', 'like', "%{$search}%")
                    ->orWhere('summary', 'like', "%{$search}%")
                    ->orWhere('content', 'like', "%{$search}%");
            });
        }

        if ($categoryId = $request->integer('category_id')) {
            $query->where('category_id', $categoryId);
        }

        if ($tagId = $request->integer('tag_id')) {
            $query->whereHas('tags', fn (Builder $builder): Builder => $builder->where('tags.id', $tagId));
        }

        $sort = $request->string('sort')->toString() ?: 'updated_at';
        $direction = $request->string('direction')->toString() === 'asc' ? 'asc' : 'desc';
        $allowedSorts = ['created_at', 'updated_at', 'title'];

        if (! in_array($sort, $allowedSorts, true)) {
            $sort = 'updated_at';
        }

        $ordered = $query->orderBy($sort, $direction);
        $perPage = max(1, min(100, $request->integer('per_page') ?: 20));

        if ($request->has('page') || $request->has('per_page')) {
            $paginated = $ordered->paginate($perPage);

            return response()->json([
                'data' => $paginated->items(),
                'meta' => [
                    'current_page' => $paginated->currentPage(),
                    'per_page' => $paginated->perPage(),
                    'last_page' => $paginated->lastPage(),
                    'total' => $paginated->total(),
                ],
            ]);
        }

        return response()->json([
            'data' => $ordered->get(),
        ]);
    }

    public function store(StorePromptRequest $request): JsonResponse
    {
        /** @var User $user */
        $user = $request->user();

        $prompt = DB::transaction(function () use ($request, $user): Prompt {
            $prompt = Prompt::query()->create([
                'user_id' => $user->id,
                'title' => $request->string('title')->toString(),
                'slug' => $this->uniqueSlug($user->id, $request->string('title')->toString()),
                'summary' => $request->input('summary'),
                'content' => $request->string('content')->toString(),
                'category_id' => $request->input('category_id'),
                'visibility' => $request->input('visibility', 'private'),
                'is_favorite' => false,
                'is_archived' => false,
            ]);

            $prompt->tags()->sync($request->input('tag_ids', []));
            $this->appendVersion($prompt);
            $this->audit($user, 'prompt.create', $prompt->id, ['title' => $prompt->title]);

            return $prompt->load(['category', 'tags']);
        });

        return response()->json([
            'message' => 'Prompt created.',
            'data' => $prompt,
        ], JsonResponse::HTTP_CREATED);
    }

    public function show(Request $request, string $promptRef): JsonResponse
    {
        $prompt = $this->findPrompt($request->user(), $promptRef)->load(['category', 'tags']);

        // Track usage
        $prompt->increment('usage_count');
        $prompt->update(['last_used_at' => now()]);

        return response()->json([
            'data' => $prompt,
        ]);
    }

    public function update(UpdatePromptRequest $request, string $promptRef): JsonResponse
    {
        /** @var User $user */
        $user = $request->user();

        $prompt = $this->findPrompt($user, $promptRef);

        DB::transaction(function () use ($request, $user, $prompt): void {
            $title = $request->input('title', $prompt->title);

            $prompt->fill([
                'title' => $title,
                'slug' => $title !== $prompt->title ? $this->uniqueSlug($user->id, $title, $prompt->id) : $prompt->slug,
                'summary' => $request->input('summary', $prompt->summary),
                'content' => $request->input('content', $prompt->content),
                'category_id' => $request->has('category_id') ? $request->input('category_id') : $prompt->category_id,
                'visibility' => $request->input('visibility', $prompt->visibility),
            ])->save();

            if ($request->has('tag_ids')) {
                $prompt->tags()->sync($request->input('tag_ids', []));
            }

            $this->appendVersion($prompt->fresh());
            $this->audit($user, 'prompt.update', $prompt->id, ['title' => $prompt->title]);
        });

        return response()->json([
            'message' => 'Prompt updated.',
            'data' => $prompt->fresh()->load(['category', 'tags']),
        ]);
    }

    public function destroy(Request $request, string $promptRef): JsonResponse
    {
        /** @var User $user */
        $user = $request->user();
        $prompt = $this->findPrompt($user, $promptRef);

        $prompt->delete();
        $this->audit($user, 'prompt.delete', $prompt->id, ['title' => $prompt->title]);

        return response()->json([
            'message' => 'Prompt deleted.',
        ]);
    }

    public function trash(Request $request): JsonResponse
    {
        /** @var User $user */
        $user = $request->user();

        return response()->json([
            'data' => Prompt::query()
                ->onlyTrashed()
                ->with(['category', 'tags'])
                ->where('user_id', $user->id)
                ->orderByDesc('deleted_at')
                ->get(),
        ]);
    }

    public function restore(Request $request, string $promptRef): JsonResponse
    {
        /** @var User $user */
        $user = $request->user();
        $prompt = $this->findPrompt($user, $promptRef, withTrashed: true, onlyTrashed: true);

        $prompt->restore();
        $this->audit($user, 'prompt.restore', $prompt->id, ['title' => $prompt->title]);

        return response()->json([
            'message' => 'Prompt restored.',
            'data' => $prompt->fresh()->load(['category', 'tags']),
        ]);
    }

    public function forceDelete(Request $request, string $promptRef): JsonResponse
    {
        /** @var User $user */
        $user = $request->user();
        $prompt = $this->findPrompt($user, $promptRef, withTrashed: true, onlyTrashed: true);

        $promptId = $prompt->id;
        $title = $prompt->title;
        $prompt->forceDelete();
        $this->audit($user, 'prompt.force_delete', $promptId, ['title' => $title]);

        return response()->json([
            'message' => 'Prompt permanently deleted.',
        ]);
    }

    public function favorite(Request $request, string $promptRef): JsonResponse
    {
        /** @var User $user */
        $user = $request->user();
        $prompt = $this->findPrompt($user, $promptRef);

        $prompt->forceFill([
            'is_favorite' => ! $prompt->is_favorite,
        ])->save();

        $this->audit($user, 'prompt.favorite', $prompt->id, ['is_favorite' => $prompt->is_favorite]);

        return response()->json([
            'message' => 'Prompt favorite flag updated.',
            'data' => $prompt,
        ]);
    }

    public function archive(Request $request, string $promptRef): JsonResponse
    {
        /** @var User $user */
        $user = $request->user();
        $prompt = $this->findPrompt($user, $promptRef);

        $prompt->forceFill([
            'is_archived' => ! $prompt->is_archived,
        ])->save();

        $this->audit($user, 'prompt.archive', $prompt->id, ['is_archived' => $prompt->is_archived]);

        return response()->json([
            'message' => 'Prompt archive flag updated.',
            'data' => $prompt,
        ]);
    }

    public function versions(Request $request, string $promptRef): JsonResponse
    {
        /** @var User $user */
        $user = $request->user();
        $prompt = $this->findPrompt($user, $promptRef);

        return response()->json([
            'data' => $prompt->versions()->orderByDesc('version_number')->get(),
        ]);
    }

    public function export(Request $request): JsonResponse
    {
        /** @var User $user */
        $user = $request->user();

        $prompts = Prompt::query()
            ->with(['category', 'tags'])
            ->where('user_id', $user->id)
            ->orderBy('updated_at', 'desc')
            ->get()
            ->map(function (Prompt $prompt): array {
                return [
                    'title' => $prompt->title,
                    'summary' => $prompt->summary,
                    'content' => $prompt->content,
                    'visibility' => $prompt->visibility,
                    'category' => $prompt->category?->name,
                    'tags' => $prompt->tags->pluck('name')->values()->all(),
                ];
            })
            ->values();

        return response()->json([
            'data' => [
                'version' => 1,
                'exported_at' => now()->toIso8601String(),
                'prompts' => $prompts,
            ],
        ]);
    }

    public function import(Request $request): JsonResponse
    {
        /** @var User $user */
        $user = $request->user();

        $request->validate([
            'prompts' => ['required', 'array', 'min:1'],
            'prompts.*.title' => ['required', 'string', 'max:255'],
            'prompts.*.content' => ['required', 'string'],
            'prompts.*.summary' => ['nullable', 'string'],
            'prompts.*.visibility' => ['nullable', 'in:private,public'],
        ]);

        $created = 0;
        $errors = [];

        DB::transaction(function () use ($request, $user, &$created, &$errors): void {
            foreach ((array) $request->input('prompts', []) as $index => $item) {
                try {
                    $prompt = Prompt::query()->create([
                        'user_id' => $user->id,
                        'title' => (string) ($item['title'] ?? ''),
                        'slug' => $this->uniqueSlug($user->id, (string) ($item['title'] ?? 'prompt')),
                        'summary' => $item['summary'] ?? null,
                        'content' => (string) ($item['content'] ?? ''),
                        'category_id' => null,
                        'visibility' => (string) ($item['visibility'] ?? 'private'),
                        'is_favorite' => false,
                        'is_archived' => false,
                    ]);

                    $this->appendVersion($prompt);
                    $this->audit($user, 'prompt.import', $prompt->id, ['title' => $prompt->title]);
                    $created++;
                } catch (\Throwable $exception) {
                    $errors[] = [
                        'index' => $index,
                        'message' => $exception->getMessage(),
                    ];
                }
            }
        });

        return response()->json([
            'message' => 'Import completed.',
            'data' => [
                'created' => $created,
                'errors' => $errors,
            ],
        ]);
    }

    private function findPrompt(User $user, string $promptRef, bool $withTrashed = false, bool $onlyTrashed = false): Prompt
    {
        $query = Prompt::query();

        if ($onlyTrashed) {
            $query->onlyTrashed();
        } elseif ($withTrashed) {
            $query->withTrashed();
        }

        return $query
            ->where('user_id', $user->id)
            ->where(function (Builder $builder) use ($promptRef): void {
                if (ctype_digit($promptRef)) {
                    $builder
                        ->where('id', (int) $promptRef)
                        ->orWhere('slug', $promptRef);
                    return;
                }

                $builder->where('slug', $promptRef);
            })
            ->firstOrFail();
    }

    private function uniqueSlug(int $userId, string $title, ?int $ignoreId = null): string
    {
        $base = Str::slug($title) ?: 'prompt';
        $slug = $base;
        $suffix = 1;

        while (
            Prompt::query()
                ->where('user_id', $userId)
                ->when($ignoreId, fn (Builder $builder): Builder => $builder->where('id', '!=', $ignoreId))
                ->where('slug', $slug)
                ->exists()
        ) {
            $suffix++;
            $slug = "{$base}-{$suffix}";
        }

        return $slug;
    }

    private function appendVersion(Prompt $prompt): void
    {
        $versionNumber = (int) ($prompt->versions()->max('version_number') ?? 0) + 1;

        PromptVersion::query()->create([
            'prompt_id' => $prompt->id,
            'content' => $prompt->content,
            'version_number' => $versionNumber,
            'created_at' => now(),
        ]);
    }

    private function audit(User $user, string $action, int $promptId, array $metadata = []): void
    {
        AuditLog::query()->create([
            'user_id' => $user->id,
            'action' => $action,
            'entity_type' => 'prompt',
            'entity_id' => $promptId,
            'metadata' => $metadata,
            'created_at' => now(),
        ]);
    }
}
