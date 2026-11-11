<?php
namespace App\Providers;
use App\Policies\SynapseRoomPolicy;
use App\Policies\SynapseUserPolicy;
use Illuminate\Support\Facades\Gate;
use Illuminate\Support\ServiceProvider;
class AppServiceProvider extends ServiceProvider {
 public function register(): void {}
 public function boot(): void {
  Gate::define('manage-users',fn($user)=>$user->hasAnyRole(['super_admin','admin']));
  Gate::define('manage-server',fn($user)=>$user->hasRole('super_admin'));
  Gate::define('view-audit-logs',fn($user)=>$user->hasAnyRole(['super_admin','admin','moderator']));
  Gate::define('view-dashboard',fn($user)=>$user->hasAnyRole(['super_admin','admin','moderator','viewer']));
  Gate::define('synapse-users-view',fn($user)=>app(SynapseUserPolicy::class)->viewAny($user));
  Gate::define('synapse-users-update',fn($user)=>app(SynapseUserPolicy::class)->update($user));
  Gate::define('synapse-users-moderate',fn($user)=>app(SynapseUserPolicy::class)->moderate($user));
  Gate::define('synapse-users-delete',fn($user)=>app(SynapseUserPolicy::class)->delete($user));
  Gate::define('synapse-rooms-view',fn($user)=>app(SynapseRoomPolicy::class)->viewAny($user));
  Gate::define('synapse-rooms-moderate',fn($user)=>app(SynapseRoomPolicy::class)->moderate($user));
  Gate::define('synapse-rooms-delete',fn($user)=>app(SynapseRoomPolicy::class)->delete($user));
 }
}
