<?php
namespace App\Http\Controllers;
use App\Services\AuditLogger;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Gate;
use Inertia\Inertia;
class QuotaController {
 public function __construct(private AuditLogger $audit) {}
 public function index(Request $r) { Gate::authorize('synapse-users-view');$quotas=DB::table('user_quotas')->orderBy('user_id')->paginate(50);return Inertia::render('Media/Index',['quotas'=>$quotas]);}
 public function store(Request $r) { Gate::authorize('synapse-users-update');$data=$r->validate(['user_id'=>'required|string|max:255','max_total_bytes'=>'required|integer|min:0','max_file_bytes'=>'required|integer|min:0','period'=>'required|in:lifetime,day,week,month']);DB::table('user_quotas')->updateOrInsert(['user_id'=>$data['user_id']],$data+['updated_by'=>$r->user()->id,'created_at'=>now(),'updated_at'=>now()]);$this->audit->record($r,'media.quota.update','matrix_user',$data['user_id'],array_diff_key($data,['user_id'=>true]));return back()->with('success','سهمیه ذخیره شد.');}
 public function destroy(Request $r,string $userID) { Gate::authorize('synapse-users-update');DB::table('user_quotas')->where('user_id',$userID)->delete();$this->audit->record($r,'media.quota.delete','matrix_user',$userID);return back();}
}
