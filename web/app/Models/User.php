<?php
namespace App\Models;
use Illuminate\Foundation\Auth\User as Authenticatable;
use Illuminate\Database\Eloquent\Relations\BelongsToMany;
use Illuminate\Notifications\Notifiable;
class User extends Authenticatable { use Notifiable; protected $fillable=['name','email','password','is_active']; protected $hidden=['password','remember_token']; protected function casts(): array { return ['email_verified_at'=>'datetime','password'=>'hashed','is_active'=>'boolean']; } public function roles(): BelongsToMany { return $this->belongsToMany(Role::class); } public function hasRole(string $role): bool { return $this->roles()->where('name',$role)->exists(); } public function hasAnyRole(array $roles): bool { return $this->roles()->whereIn('name',$roles)->exists(); } }
