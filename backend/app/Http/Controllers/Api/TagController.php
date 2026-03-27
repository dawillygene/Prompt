<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use App\Http\Requests\StoreTagRequest;
use App\Models\AuditLog;
use App\Models\Tag;
use App\Models\User;
use Illuminate\Database\Eloquent\Builder;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Str;

class TagController extends Controller
{
    public function index(Request $request): JsonResponse
    {
        /** @var User $user */
        $user = $request->user();

        return response()->json([
            'data' => Tag::query()
                ->where('user_id', $user->id)
                ->orderBy('name')
                ->get(),
        ]);
    }

    public function store(StoreTagRequest $request): JsonResponse
    {
        /** @var User $user */
        $user = $request->user();

        $tag = Tag::query()->create([
            'user_id' => $user->id,
            'name' => $request->string('name')->toString(),
            'slug' => $this->uniqueSlug($user->id, $request->string('name')->toString()),
            'description' => $request->input('description'),
        ]);

        AuditLog::query()->create([
            'user_id' => $user->id,
            'action' => 'tag.create',
            'entity_type' => 'tag',
            'entity_id' => $tag->id,
            'metadata' => ['name' => $tag->name],
            'created_at' => now(),
        ]);

        return response()->json([
            'message' => 'Tag created.',
            'data' => $tag,
        ], JsonResponse::HTTP_CREATED);
    }

    public function update(Request $request, string $tagRef): JsonResponse
    {
        /** @var User $user */
        $user = $request->user();

        $request->validate([
            'name' => ['sometimes', 'string', 'max:100'],
            'description' => ['sometimes', 'nullable', 'string'],
        ]);

        $tag = $this->findTag($user, $tagRef);
        $name = $request->input('name', $tag->name);

        $tag->fill([
            'name' => $name,
            'slug' => $name !== $tag->name
                ? $this->uniqueSlug($user->id, $name, $tag->id)
                : $tag->slug,
            'description' => $request->has('description') ? $request->input('description') : $tag->description,
        ])->save();

        AuditLog::query()->create([
            'user_id' => $user->id,
            'action' => 'tag.update',
            'entity_type' => 'tag',
            'entity_id' => $tag->id,
            'metadata' => ['name' => $tag->name],
            'created_at' => now(),
        ]);

        return response()->json([
            'message' => 'Tag updated.',
            'data' => $tag->fresh(),
        ]);
    }

    public function destroy(Request $request, string $tagRef): JsonResponse
    {
        /** @var User $user */
        $user = $request->user();

        $tag = $this->findTag($user, $tagRef);
        $tagId = $tag->id;
        $name = $tag->name;
        $tag->delete();

        AuditLog::query()->create([
            'user_id' => $user->id,
            'action' => 'tag.delete',
            'entity_type' => 'tag',
            'entity_id' => $tagId,
            'metadata' => ['name' => $name],
            'created_at' => now(),
        ]);

        return response()->json([
            'message' => 'Tag deleted.',
        ]);
    }

    private function findTag(User $user, string $tagRef): Tag
    {
        return Tag::query()
            ->where('user_id', $user->id)
            ->where(function (Builder $builder) use ($tagRef): void {
                if (ctype_digit($tagRef)) {
                    $builder
                        ->where('id', (int) $tagRef)
                        ->orWhere('slug', $tagRef);
                    return;
                }

                $builder->where('slug', $tagRef);
            })
            ->firstOrFail();
    }

    private function uniqueSlug(int $userId, string $name, ?int $ignoreId = null): string
    {
        $base = Str::slug($name) ?: 'tag';
        $slug = $base;
        $suffix = 1;

        while (
            Tag::query()
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
}
