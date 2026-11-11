<?php
namespace App\Http\Controllers;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Gate;
use Inertia\Inertia;
class MetricsController {
 public function index(){Gate::authorize('view-dashboard');$latest=DB::table('system_metrics')->orderByDesc('sampled_at')->first();$history=DB::table('system_metrics')->orderByDesc('sampled_at')->limit(120)->get()->reverse()->values();$thresholds=DB::table('metric_thresholds')->pluck('value','name');return Inertia::render('Metrics/Index',['latest'=>$latest,'history'=>$history,'thresholds'=>$thresholds]);}
 public function updateThresholds(Request $r){Gate::authorize('manage-server');$data=$r->validate(['cpu_percent'=>'required|numeric|min:1|max:100','ram_percent'=>'required|numeric|min:1|max:100','disk_percent'=>'required|numeric|min:1|max:100']);foreach($data as $name=>$value)DB::table('metric_thresholds')->updateOrInsert(['name'=>$name],['value'=>$value,'updated_at'=>now(),'created_at'=>now()]);return back()->with('success','آستانه‌های هشدار ذخیره شد.');}
}
