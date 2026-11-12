<?php
namespace App\Http\Controllers;
use App\Services\AuditLogger;
use App\Services\CoreClient;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Gate;
use Inertia\Inertia;
class QuotaController {
 public function __construct(private AuditLogger $audit,private CoreClient $core) {}
 public function index(Request $r) {
  Gate::authorize('synapse-users-view');
  $quotas=DB::table('user_quotas')->orderBy('user_id')->paginate(50);
  $quotaRows=DB::table('user_quotas')->get()->keyBy('user_id');
  try {
   $stats=$this->core->request('GET','/api/v1/media/stats',['limit'=>100,'order_by'=>'media_length','dir'=>'b']);
   $rawUsers=$stats['users']??[];
   $mediaStatsError=null;
  } catch(\\Throwable $e) {
   $rawUsers=[];
   $mediaStatsError='آمار مدیای Synapse در دسترس نیست.';
  }
  $seen=[];
  $mediaUsers=[];
  $usage=[];
  foreach($rawUsers as $item) {
   $id=(string)($item['user_id']??'');
   if($id==='') continue;
   $quota=$quotaRows->get($id);
   $bytes=(int)($item['media_length']??0);
   $mediaUsers[]=[
    'user_id'=>$id,'displayname'=>$item['displayname']??null,
    'media_count'=>(int)($item['media_count']??0),'media_length'=>$bytes,
    'max_total_bytes'=>(int)($quota->max_total_bytes??0),
    'max_file_bytes'=>(int)($quota->max_file_bytes??0),'period'=>$quota->period??'lifetime'
   ];
   $usage[$id]=['total_bytes'=>$bytes];
   $seen[$id]=true;
  }
  foreach($quotaRows as $id=>$quota) {
   if(isset($seen[$id])) continue;
   $mediaUsers[]=['user_id'=>$id,'displayname'=>null,'media_count'=>0,'media_length'=>0,
    'max_total_bytes'=>(int)$quota->max_total_bytes,'max_file_bytes'=>(int)$quota->max_file_bytes,'period'=>$quota->period];
   $usage[$id]=['total_bytes'=>0];
  }
  return Inertia::render('Media/Index',[
   'quotas'=>$quotas,'usage'=>$usage,'mediaUsers'=>$mediaUsers,
   'totalQuota'=>(int)array_sum(array_map(fn($u)=>$u['max_total_bytes'],$mediaUsers)),
   'mediaStatsError'=>$mediaStatsError
  ]);
 }
 public function store(Request $r) { Gate::authorize('synapse-users-update');$data=$r->validate(['user_id'=>'required|string|max:255','max_total_bytes'=>'required|integer|min:0','max_file_bytes'=>'required|integer|min:0','period'=>'required|in:lifetime,day,week,month']);DB::table('user_quotas')->updateOrInsert(['user_id'=>$data['user_id']],$data+['updated_by'=>$r->user()->id,'created_at'=>now(),'updated_at'=>now()]);$this->audit->record($r,'media.quota.update','matrix_user',$data['user_id'],array_diff_key($data,['user_id'=>true]));return back()->with('success','سهمیه ذخیره شد.');}
 public function destroy(Request $r,string $userID) { Gate::authorize('synapse-users-update');DB::table('user_quotas')->where('user_id',$userID)->delete();$this->audit->record($r,'media.quota.delete','matrix_user',$userID);return back();}
}
