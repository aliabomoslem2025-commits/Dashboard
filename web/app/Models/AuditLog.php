<?php
namespace App\Models;
use Illuminate\Database\Eloquent\Model;
class AuditLog extends Model { public $timestamps=false; protected $fillable=['user_id','action','target_type','target_id','payload','ip_address','created_at']; protected function casts(): array { return ['payload'=>'array','created_at'=>'datetime']; } }
