<?php
namespace App\Http\Controllers;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Gate;
use Inertia\Inertia;
class ActivityController {
 public function index(Request $r){Gate::authorize('synapse-users-view');$q=DB::table('user_activity')->when($r->filled('category'),fn($q)=>$q->where('category',$r->string('category')))->when($r->filled('search'),fn($q)=>$q->where('user_id','ilike','%'.$r->string('search').'%'))->orderByDesc('message_count')->paginate(50)->withQueryString();$stats=DB::table('user_activity')->selectRaw('category,count(*) as total')->groupBy('category')->pluck('total','category');$thresholds=DB::table('activity_thresholds')->first();$history=DB::table('activity_history')->orderBy('sampled_at')->limit(100)->get();return Inertia::render('Activity/Index',['users'=>$q,'stats'=>$stats,'thresholds'=>$thresholds,'history'=>$history,'filters'=>$r->only(['category','search'])]);}
 public function updateThresholds(Request $r){Gate::authorize('manage-server');$data=$r->validate(['heavy_min_messages'=>'required|integer|min:1','active_min_messages'=>'required|integer|min:1','low_active_min_messages'=>'required|integer|min:0','dormant_days'=>'required|integer|min:1','suspicious_new_account_days'=>'required|integer|min:0','suspicious_messages_per_hour'=>'required|integer|min:1']);DB::table('activity_thresholds')->updateOrInsert(['id'=>1],$data+['updated_by'=>$r->user()->id,'created_at'=>now(),'updated_at'=>now()]);return back()->with('success','آستانه‌های فعالیت ذخیره شد.');}
}
