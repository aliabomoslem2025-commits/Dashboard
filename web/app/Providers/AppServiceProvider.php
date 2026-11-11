<?php
namespace App\Providers;
use App\Models\Broadcast;
use App\Models\SynapseRoom;
use App\Models\SynapseUser;
use App\Policies\BroadcastPolicy;
use App\Policies\SynapseRoomPolicy;
use App\Policies\SynapseUserPolicy;
use Illuminate\Support\Facades\Gate;
use Illuminate\Support\ServiceProvider;
class AppServiceProvider extends ServiceProvider {
 public function register(): void {}
 public function boot(): void {
  Gate::policy(SynapseUser::class,SynapseUserPolicy::class);
  Gate::policy(SynapseRoom::class,SynapseRoomPolicy::class);
  Gate::policy(Broadcast::class,BroadcastPolicy::class);
  Gate::define('synapse-users-view',[SynapseUserPolicy::class,'viewAny']);
  Gate::define('synapse-users-update',fn($user)=>$user->hasAnyRole(['super_admin','admin']));
  Gate::define('synapse-users-moderate',fn($user)=>$user->hasAnyRole(['super_admin','admin','moderator']));
  Gate::define('synapse-users-delete',fn($user)=>$user->hasRole('super_admin'));
  Gate::define('synapse-rooms-view',[SynapseRoomPolicy::class,'viewAny']);
  Gate::define('synapse-rooms-update',fn($user)=>$user->hasAnyRole(['super_admin','admin']));
  Gate::define('synapse-rooms-moderate',fn($user)=>$user->hasAnyRole(['super_admin','admin','moderator']));
  Gate::define('synapse-rooms-delete',fn($user)=>$user->hasAnyRole(['super_admin','admin']));
  Gate::define('broadcasts-view',[BroadcastPolicy::class,'viewAny']);
  Gate::define('broadcasts-create',[BroadcastPolicy::class,'create']);
  Gate::define('manage-server',fn($user)=>$user->hasRole('super_admin'));
  Gate::define('view-audit-logs',fn($user)=>$user->hasAnyRole(['super_admin','admin','moderator']));
  Gate::define('view-dashboard',fn($user)=>$user->hasAnyRole(['super_admin','admin','moderator','viewer']));
 }
}
