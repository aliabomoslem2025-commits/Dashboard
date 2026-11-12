<?php
namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\HasMany;

class Broadcast extends Model
{
    protected $table = 'broadcasts';
    protected $guarded = ['id'];

    protected $casts = [
        'started_at' => 'datetime',
        'completed_at' => 'datetime',
        'created_at' => 'datetime',
        'updated_at' => 'datetime',
    ];

    public function recipients(): HasMany
    {
        return $this->hasMany(BroadcastRecipient::class, 'broadcast_id');
    }
}
