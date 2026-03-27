<?php

namespace App\Http\Controllers\Api;

use App\Http\Controllers\Controller;
use App\Http\Requests\LoginRequest;
use App\Http\Requests\RegisterRequest;
use App\Models\AuditLog;
use App\Models\User;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Support\Str;

class AuthController extends Controller
{
    public function register(RegisterRequest $request): JsonResponse
    {
        $plainToken = Str::random(64);

        $user = User::query()->create([
            'name' => $request->string('name')->toString(),
            'email' => $request->string('email')->toString(),
            'password' => $request->string('password')->toString(),
            'api_token' => hash('sha256', $plainToken),
        ]);

        $this->audit($user, 'auth.register', 'user', $user->id);

        return response()->json([
            'message' => 'Registration successful.',
            'token' => $plainToken,
            'user' => $user,
        ], JsonResponse::HTTP_CREATED);
    }

    public function login(LoginRequest $request): JsonResponse
    {
        $user = User::query()
            ->where('email', $request->string('email')->toString())
            ->first();

        if (! $user || ! password_verify($request->string('password')->toString(), $user->password)) {
            return response()->json([
                'message' => 'Invalid credentials.',
            ], JsonResponse::HTTP_UNAUTHORIZED);
        }

        $plainToken = Str::random(64);
        $user->forceFill([
            'api_token' => hash('sha256', $plainToken),
        ])->save();

        $this->audit($user, 'auth.login', 'user', $user->id);

        return response()->json([
            'message' => 'Login successful.',
            'token' => $plainToken,
            'user' => $user,
        ]);
    }

    public function logout(Request $request): JsonResponse
    {
        /** @var User $user */
        $user = $request->user();

        $user->forceFill([
            'api_token' => null,
        ])->save();

        $this->audit($user, 'auth.logout', 'user', $user->id);

        return response()->json([
            'message' => 'Logout successful.',
        ]);
    }

    public function me(Request $request): JsonResponse
    {
        return response()->json([
            'user' => $request->user(),
        ]);
    }

    private function audit(User $user, string $action, string $entityType, int $entityId, array $metadata = []): void
    {
        AuditLog::query()->create([
            'user_id' => $user->id,
            'action' => $action,
            'entity_type' => $entityType,
            'entity_id' => $entityId,
            'metadata' => $metadata,
            'created_at' => now(),
        ]);
    }
}
