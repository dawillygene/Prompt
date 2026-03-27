<?php

namespace App\Models;

use Illuminate\Database\Eloquent\Model;

class SyncQueue extends Model
{
    protected $table = 'sync_queues';

    #[\Illuminate\Database\Eloquent\Casts\Attributes\Fillable(['user_id', 'action', 'entity_type', 'entity_id', 'payload', 'status', 'conflict_with_id', 'conflict_type'])]

    protected $casts = [
        'payload' => 'array',
    ];

    public function user()
    {
        return $this->belongsTo(User::class);
    }

    // Scope: get pending syncs for a user
    public function scopePending($query)
    {
        return $query->where('status', 'pending');
    }

    // Scope: get conflicts for a user
    public function scopeConflicts($query)
    {
        return $query->where('status', 'conflict');
    }

    // Scope: get synced items for a user
    public function scopeSynced($query)
    {
        return $query->where('status', 'synced');
    }
}
