<?php
namespace App\Providers;
use Illuminate\Support\Facades\Gate;
use Illuminate\Support\ServiceProvider;
class AppServiceProvider extends ServiceProvider { public function register(): void {} public function boot(): void { Gate::define('manage-users',fn($user)=>$user->hasAnyRole(['super_admin','admin'])); Gate::define('manage-server',fn($user)=>$user->hasRole('super_admin')); Gate::define('view-audit-logs',fn($user)=>$user->hasAnyRole(['super_admin','admin','moderator'])); Gate::define('view-dashboard',fn($user)=>$user->hasAnyRole(['super_admin','admin','moderator','viewer'])); } }
