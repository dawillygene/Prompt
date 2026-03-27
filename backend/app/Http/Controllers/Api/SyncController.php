<?php

namespace App\Http\Controllers\Api;

use App\Models\Prompt;
use App\Models\SyncQueue;
use Illuminate\Http\JsonResponse;
use Illuminate\Http\Request;
use Illuminate\Routing\Controller;

class SyncController extends Controller
{
    /**
     * Get sync status: pending queue items and conflicts for the user
     */
    public function status(Request $request): JsonResponse
    {
        $user = $request->user();

        $pending = SyncQueue::where('user_id', $user->id)
            ->where('status', 'pending')
            ->get();

        $conflicts = SyncQueue::where('user_id', $user->id)
            ->where('status', 'conflict')
            ->get();

        return response()->json([
            'data' => [
                'pending_count' => $pending->count(),
                'conflict_count' => $conflicts->count(),
                'pending_items' => $pending,
                'conflicts' => $conflicts,
            ],
        ]);
    }

    /**
     * Submit sync queue items and detect conflicts
     * POST /api/sync
     * Request body: { queue_items: [...] }
     */
    public function sync(Request $request): JsonResponse
    {
        $user = $request->user();
        $incomingItems = $request->input('queue_items', []);

        $conflicts = [];
        $synced = 0;

        foreach ($incomingItems as $item) {
            $action = $item['action'] ?? null;
            $entityType = $item['entity_type'] ?? null;
            $entityId = $item['entity_id'] ?? null;
            $payload = $item['payload'] ?? [];

            // Find the server entity to check for conflicts
            $serverEntity = null;
            if ($entityType === 'prompt' && $entityId) {
                $serverEntity = Prompt::withTrashed()
                    ->where('user_id', $user->id)
                    ->find($entityId);
            }

            // Detect conflicts: if entity exists on server but has newer timestamp
            if ($serverEntity && $action !== 'create') {
                $clientUpdatedAt = strtotime($payload['updated_at'] ?? '1970-01-01');
                $serverUpdatedAt = strtotime($serverEntity->updated_at);

                if ($serverUpdatedAt > $clientUpdatedAt) {
                    // Server is newer - conflict
                    $queueItem = SyncQueue::create([
                        'user_id' => $user->id,
                        'action' => $action,
                        'entity_type' => $entityType,
                        'entity_id' => $entityId,
                        'payload' => $payload,
                        'status' => 'conflict',
                        'conflict_with_id' => $entityId,
                        'conflict_type' => 'newer_on_server',
                    ]);

                    $conflicts[] = [
                        'queue_id' => $queueItem->id,
                        'entity_type' => $entityType,
                        'entity_id' => $entityId,
                        'action' => $action,
                        'conflict_type' => 'newer_on_server',
                        'server_updated_at' => $serverEntity->updated_at,
                        'client_payload' => $payload,
                        'server_data' => $serverEntity->toArray(),
                    ];

                    continue;
                }
            }

            // No conflict - process the sync
            $this->processSyncItem($user->id, $action, $entityType, $entityId, $payload);

            SyncQueue::create([
                'user_id' => $user->id,
                'action' => $action,
                'entity_type' => $entityType,
                'entity_id' => $entityId,
                'payload' => $payload,
                'status' => 'synced',
            ]);

            $synced++;
        }

        return response()->json([
            'data' => [
                'synced' => $synced,
                'conflicts' => $conflicts,
            ],
        ], $conflicts ? 409 : 200); // 409 Conflict if there are conflicts
    }

    /**
     * Resolve a conflict - user chooses which version to keep
     * POST /api/sync/resolve
     * Request body: { queue_id, strategy: 'keep_local' | 'keep_server' | 'merge' }
     */
    public function resolve(Request $request): JsonResponse
    {
        $user = $request->user();
        $queueId = $request->input('queue_id');
        $strategy = $request->input('strategy'); // 'keep_local', 'keep_server', 'merge'

        $queueItem = SyncQueue::where('id', $queueId)
            ->where('user_id', $user->id)
            ->where('status', 'conflict')
            ->firstOrFail();

        if ($strategy === 'keep_local') {
            // Apply the local change
            $this->processSyncItem(
                $user->id,
                $queueItem->action,
                $queueItem->entity_type,
                $queueItem->entity_id,
                $queueItem->payload
            );
        } elseif ($strategy === 'keep_server') {
            // Discard local change, keep server version
            // No action needed - just mark as synced
        } elseif ($strategy === 'merge') {
            // For merge: apply the local change (simplified merge strategy)
            $this->processSyncItem(
                $user->id,
                $queueItem->action,
                $queueItem->entity_type,
                $queueItem->entity_id,
                $queueItem->payload
            );
        }

        $queueItem->update(['status' => 'synced']);

        return response()->json([
            'message' => 'Conflict resolved.',
            'data' => [
                'queue_id' => $queueId,
                'strategy' => $strategy,
            ],
        ]);
    }

    /**
     * Apply a sync item to the database
     */
    private function processSyncItem(int $userId, string $action, string $entityType, int $entityId, array $payload): void
    {
        if ($entityType === 'prompt') {
            match ($action) {
                'create' => Prompt::create(array_merge($payload, ['user_id' => $userId])),
                'update' => Prompt::where('id', $entityId)
                    ->where('user_id', $userId)
                    ->update($payload),
                'delete' => Prompt::where('id', $entityId)
                    ->where('user_id', $userId)
                    ->forceDelete(),
                default => null,
            };
        }
    }
}
