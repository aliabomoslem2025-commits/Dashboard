<?php
namespace App\Models;

use Illuminate\Database\Eloquent\Model;
use Illuminate\Database\Eloquent\Relations\BelongsTo;

class BroadcastRecipient extends Model
{
    protected $table = 'broadcast_recipients';
    protected $guarded = ['id'];

    protected $casts = [
        'sent_at' => 'datetime',
        'created_at' => 'datetime',
        'updated_at' => 'datetime',
        'attempts' => 'integer',
    ];

    public function broadcast(): BelongsTo
    {
        return $this->belongsTo(Broadcast::class, 'broadcast_id');
    }
}
