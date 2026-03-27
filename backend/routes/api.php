<?php

use App\Http\Controllers\Api\AuthController;
use App\Http\Controllers\Api\CategoryController;
use App\Http\Controllers\Api\PromptController;
use App\Http\Controllers\Api\SyncController;
use App\Http\Controllers\Api\TagController;
use Illuminate\Support\Facades\Route;

Route::post('/register', [AuthController::class, 'register']);
Route::post('/login', [AuthController::class, 'login']);

Route::middleware('auth.token')->group(function (): void {
    Route::post('/logout', [AuthController::class, 'logout']);
    Route::get('/me', [AuthController::class, 'me']);

    Route::get('/prompts', [PromptController::class, 'index']);
    Route::get('/prompts/trash', [PromptController::class, 'trash']);
    Route::get('/export', [PromptController::class, 'export']);
    Route::post('/import', [PromptController::class, 'import']);
    Route::post('/prompts', [PromptController::class, 'store']);
    Route::get('/prompts/{promptRef}', [PromptController::class, 'show']);
    Route::put('/prompts/{promptRef}', [PromptController::class, 'update']);
    Route::delete('/prompts/{promptRef}', [PromptController::class, 'destroy']);
    Route::post('/prompts/{promptRef}/restore', [PromptController::class, 'restore']);
    Route::delete('/prompts/{promptRef}/force', [PromptController::class, 'forceDelete']);
    Route::post('/prompts/{promptRef}/favorite', [PromptController::class, 'favorite']);
    Route::post('/prompts/{promptRef}/archive', [PromptController::class, 'archive']);
    Route::get('/prompts/{promptRef}/versions', [PromptController::class, 'versions']);

    Route::get('/categories', [CategoryController::class, 'index']);
    Route::post('/categories', [CategoryController::class, 'store']);
    Route::put('/categories/{categoryRef}', [CategoryController::class, 'update']);
    Route::delete('/categories/{categoryRef}', [CategoryController::class, 'destroy']);

    Route::get('/tags', [TagController::class, 'index']);
    Route::post('/tags', [TagController::class, 'store']);
    Route::put('/tags/{tagRef}', [TagController::class, 'update']);
    Route::delete('/tags/{tagRef}', [TagController::class, 'destroy']);

    // Sync queue and conflict resolution
    Route::get('/sync/status', [SyncController::class, 'status']);
    Route::post('/sync', [SyncController::class, 'sync']);
    Route::post('/sync/resolve', [SyncController::class, 'resolve']);
});
