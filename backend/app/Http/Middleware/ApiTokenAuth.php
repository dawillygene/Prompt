<?php

namespace App\Http\Middleware;

use App\Models\User;
use Closure;
use Illuminate\Http\Request;
use Symfony\Component\HttpFoundation\Response;

class ApiTokenAuth
{
    public function handle(Request $request, Closure $next): Response
    {
        $plainToken = $request->bearerToken();

        if (! $plainToken) {
            return response()->json([
                'message' => 'Missing bearer token.',
            ], Response::HTTP_UNAUTHORIZED);
        }

        $user = User::query()
            ->where('api_token', hash('sha256', $plainToken))
            ->first();

        if (! $user) {
            return response()->json([
                'message' => 'Invalid bearer token.',
            ], Response::HTTP_UNAUTHORIZED);
        }

        $request->setUserResolver(fn (): User => $user);

        return $next($request);
    }
}
