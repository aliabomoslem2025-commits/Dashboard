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
  Gate::define('synapse-users-update',[SynapseUserPolicy::class,'update']);
  Gate::define('synapse-users-moderate',[SynapseUserPolicy::class,'moderate']);
  Gate::define('synapse-users-delete',[SynapseUserPolicy::class,'delete']);
  Gate::define('synapse-rooms-view',[SynapseRoomPolicy::class,'viewAny']);
  Gate::define('synapse-rooms-update',[SynapseRoomPolicy::class,'update']);
  Gate::define('synapse-rooms-moderate',[SynapseRoomPolicy::class,'moderate']);
  Gate::define('synapse-rooms-delete',[SynapseRoomPolicy::class,'delete']);
  Gate::define('broadcasts-view',[BroadcastPolicy::class,'viewAny']);
  Gate::define('broadcasts-create',[BroadcastPolicy::class,'create']);
  Gate::define('manage-server',fn($user)=>$user->hasRole('super_admin'));
  Gate::define('view-audit-logs',fn($user)=>$user->hasAnyRole(['super_admin','admin','moderator']));
  Gate::define('view-dashboard',fn($user)=>$user->hasAnyRole(['super_admin','admin','moderator','viewer']));
 }
}
