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
  Gate::define('synapse-users-view',fn($user)=>app(SynapseUserPolicy::class)->viewAny($user));
  Gate::define('synapse-users-update',fn($user)=>app(SynapseUserPolicy::class)->update($user));
  Gate::define('synapse-users-moderate',fn($user)=>app(SynapseUserPolicy::class)->moderate($user));
  Gate::define('synapse-users-delete',fn($user)=>app(SynapseUserPolicy::class)->delete($user));
  Gate::define('synapse-rooms-view',fn($user)=>app(SynapseRoomPolicy::class)->viewAny($user));
  Gate::define('synapse-rooms-update',fn($user)=>app(SynapseRoomPolicy::class)->update($user));
  Gate::define('synapse-rooms-moderate',fn($user)=>app(SynapseRoomPolicy::class)->moderate($user));
  Gate::define('synapse-rooms-delete',fn($user)=>app(SynapseRoomPolicy::class)->delete($user));
  Gate::define('broadcasts-view',fn($user)=>app(BroadcastPolicy::class)->viewAny($user));
  Gate::define('broadcasts-create',fn($user)=>app(BroadcastPolicy::class)->create($user));
  Gate::define('manage-server',fn($user)=>$user->hasRole('super_admin'));
  Gate::define('view-audit-logs',fn($user)=>$user->hasAnyRole(['super_admin','admin','moderator']));
  Gate::define('view-dashboard',fn($user)=>$user->hasAnyRole(['super_admin','admin','moderator','viewer']));
 }
}
