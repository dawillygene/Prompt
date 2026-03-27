<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use App\Http\Requests\StoreCategoryRequest;
use App\Models\AuditLog;
use App\Models\Category;
use App\Models\User;
use Illuminate\Database\Eloquent\Builder;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Str;

class CategoryController extends Controller
{
    public function index(Request $request): JsonResponse
    {
        /** @var User $user */
        $user = $request->user();

        return response()->json([
            'data' => Category::query()
                ->where('user_id', $user->id)
                ->orderBy('name')
                ->get(),
        ]);
    }

    public function store(StoreCategoryRequest $request): JsonResponse
    {
        /** @var User $user */
        $user = $request->user();

        $category = Category::query()->create([
            'user_id' => $user->id,
            'name' => $request->string('name')->toString(),
            'slug' => $this->uniqueSlug($user->id, $request->string('name')->toString()),
            'description' => $request->input('description'),
        ]);

        AuditLog::query()->create([
            'user_id' => $user->id,
            'action' => 'category.create',
            'entity_type' => 'category',
            'entity_id' => $category->id,
            'metadata' => ['name' => $category->name],
            'created_at' => now(),
        ]);

        return response()->json([
            'message' => 'Category created.',
            'data' => $category,
        ], JsonResponse::HTTP_CREATED);
    }

    public function update(Request $request, string $categoryRef): JsonResponse
    {
        /** @var User $user */
        $user = $request->user();

        $request->validate([
            'name' => ['sometimes', 'string', 'max:100'],
            'description' => ['sometimes', 'nullable', 'string'],
        ]);

        $category = $this->findCategory($user, $categoryRef);
        $newName = $request->input('name', $category->name);

        $category->fill([
            'name' => $newName,
            'slug' => $newName !== $category->name
                ? $this->uniqueSlug($user->id, $newName, $category->id)
                : $category->slug,
            'description' => $request->has('description') ? $request->input('description') : $category->description,
        ])->save();

        AuditLog::query()->create([
            'user_id' => $user->id,
            'action' => 'category.update',
            'entity_type' => 'category',
            'entity_id' => $category->id,
            'metadata' => ['name' => $category->name],
            'created_at' => now(),
        ]);

        return response()->json([
            'message' => 'Category updated.',
            'data' => $category->fresh(),
        ]);
    }

    public function destroy(Request $request, string $categoryRef): JsonResponse
    {
        /** @var User $user */
        $user = $request->user();

        $category = $this->findCategory($user, $categoryRef);
        $categoryId = $category->id;
        $name = $category->name;
        $category->delete();

        AuditLog::query()->create([
            'user_id' => $user->id,
            'action' => 'category.delete',
            'entity_type' => 'category',
            'entity_id' => $categoryId,
            'metadata' => ['name' => $name],
            'created_at' => now(),
        ]);

        return response()->json([
            'message' => 'Category deleted.',
        ]);
    }

    private function findCategory(User $user, string $categoryRef): Category
    {
        return Category::query()
            ->where('user_id', $user->id)
            ->where(function (Builder $builder) use ($categoryRef): void {
                if (ctype_digit($categoryRef)) {
                    $builder
                        ->where('id', (int) $categoryRef)
                        ->orWhere('slug', $categoryRef);
                    return;
                }

                $builder->where('slug', $categoryRef);
            })
            ->firstOrFail();
    }

    private function uniqueSlug(int $userId, string $name, ?int $ignoreId = null): string
    {
        $base = Str::slug($name) ?: 'category';
        $slug = $base;
        $suffix = 1;

        while (
            Category::query()
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
