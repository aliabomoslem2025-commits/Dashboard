<?php
namespace App\Http\Controllers;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Gate;
use Inertia\Inertia;
class DashboardController {
 public function index(){Gate::authorize('view-dashboard');$latest=DB::table('system_metrics')->orderByDesc('sampled_at')->first();$activity=DB::table('user_activity')->selectRaw('category,count(*) as total')->groupBy('category')->pluck('total','category');$queued=DB::table('broadcasts')->whereIn('status',['queued','sending'])->count();$history=DB::table('system_metrics')->orderByDesc('sampled_at')->limit(30)->get()->reverse()->values();$thresholds=DB::table('metric_thresholds')->pluck('value','name');return Inertia::render('Dashboard',['latest'=>$latest,'activity'=>$activity,'queuedBroadcasts'=>$queued,'history'=>$history,'thresholds'=>$thresholds]);}
}
