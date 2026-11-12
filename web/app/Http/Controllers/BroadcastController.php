<?php
namespace App\Http\Controllers;
use App\Services\AuditLogger;
use App\Services\CoreClient;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Gate;
use Illuminate\Support\Str;
use Inertia\Inertia;
class BroadcastController {
 public function __construct(private AuditLogger $audit,private CoreClient $core) {}
 private function allActiveUsers(): array {$users=[];$from='0';for($page=0;$page<10000;$page++){$data=$this->core->request('GET','/api/v1/users',['from'=>$from,'limit'=>500]);foreach(($data['users']??[]) as $user){if(!empty($user['name'])&&empty($user['deactivated']))$users[]=$user['name'];}$next=$data['next_token']??null;if($next===null||$next==='')break;$from=(string)$next;}return array_values(array_unique($users));}
 public function sendOne(Request $r){
  Gate::authorize('broadcasts-create');
  $data=$r->validate(['user_id'=>['required','string','regex:/^@[^\\s]+:[^\\s]+$/','max:255'],'message'=>['required','string','max:4000']]);
  $txnId=(string) Str::uuid();
  $payload=['user_id'=>$data['user_id'],'content'=>['msgtype'=>'m.text','body'=>$data['message']]];
  $result=$this->core->request('PUT','/api/v1/notices/'.rawurlencode($txnId),[],$payload);
  $this->audit->record($r,'broadcast.single.sent','matrix_user',$data['user_id'],['event_id'=>$result['event_id']??null,'transaction_id'=>$txnId]);
  return back()->with('success','پیام اطلاع‌رسانی برای کاربر ارسال شد.');
 }
 public function index(){Gate::authorize('broadcasts-view');$items=DB::table('broadcasts')->orderByDesc('id')->paginate(30);$items->getCollection()->transform(function($b){$b->total=DB::table('broadcast_recipients')->where('broadcast_id',$b->id)->count();$b->sent=DB::table('broadcast_recipients')->where('broadcast_id',$b->id)->where('status','sent')->count();return $b;});return Inertia::render('Broadcast/Index',['broadcasts'=>$items]);}
 public function store(Request $r){Gate::authorize('broadcasts-create');$data=$r->validate(['message'=>'required|string|max:4000','audience'=>'required|in:all,category,selected','activity_category'=>'nullable|in:heavy_active,active,low_active,dormant,suspicious','recipients'=>'nullable|array|max:5000','recipients.*'=>['required','string','regex:/^@[^\\\\s]+:[^\\\\s]+$/','max:255'],'throttle_ms'=>'required|integer|min:100|max:60000','max_retries'=>'required|integer|min:0|max:10','confirmed'=>'accepted']);$recipients=match($data['audience']){'selected'=>array_values(array_unique($data['recipients']??[])),'category'=>DB::table('user_activity')->where('category',$data['activity_category'])->pluck('user_id')->all(),default=>$this->allActiveUsers()};if(count($recipients)===0)return back()->withErrors(['recipients'=>'هیچ گیرنده‌ای برای این انتخاب پیدا نشد.']);$id=DB::table('broadcasts')->insertGetId(['message'=>$data['message'],'msgtype'=>'m.text','audience'=>$data['audience'],'activity_category'=>$data['activity_category']??null,'throttle_ms'=>$data['throttle_ms'],'max_retries'=>$data['max_retries'],'status'=>'queued','created_by'=>$r->user()->id,'created_at'=>now(),'updated_at'=>now()]);foreach(array_chunk($recipients,500) as $chunk){DB::table('broadcast_recipients')->insert(array_map(fn($user)=>['broadcast_id'=>$id,'user_id'=>$user,'status'=>'pending','created_at'=>now(),'updated_at'=>now()],$chunk));}$this->audit->record($r,'broadcast.create','broadcast',(string)$id,['audience'=>$data['audience'],'activity_category'=>$data['activity_category']??null,'recipient_count'=>count($recipients),'throttle_ms'=>$data['throttle_ms']]);return redirect()->route('broadcasts.show',$id)->with('success','پیام در صف ارسال قرار گرفت.');}
 public function show(int $id){Gate::authorize('broadcasts-view');$broadcast=DB::table('broadcasts')->find($id);abort_unless($broadcast,404);$recipients=DB::table('broadcast_recipients')->where('broadcast_id',$id)->orderBy('id')->paginate(100);$progress=DB::table('broadcast_recipients')->where('broadcast_id',$id)->selectRaw("COUNT(*) as total, SUM(CASE WHEN status='sent' THEN 1 ELSE 0 END) as sent, SUM(CASE WHEN status='failed' THEN 1 ELSE 0 END) as failed")->first();return Inertia::render('Broadcast/Show',['broadcast'=>$broadcast,'recipients'=>$recipients,'progress'=>$progress]);}
}
