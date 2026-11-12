<?php
namespace App\Http\Controllers;
use App\Models\SynapseUser;
use App\Services\AuditLogger;
use App\Services\CoreClient;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Gate;
use Inertia\Inertia;
class SynapseUserController {
 public function __construct(private CoreClient $core,private AuditLogger $audit) {}
 private function id(string $id): string { return rawurlencode($id); }
 private function supportsLocked(): bool { try{$v=$this->core->request('GET','/api/v1/server/version');return version_compare((string)($v['server_version']??''),'1.93.0','>=');}catch(\Throwable){return false;} }
 private function optional(string $method,string $path,array $query=[]): array { try{return $this->core->request($method,$path,$query);}catch(\Throwable $e){return ['_unavailable'=>true,'_message'=>$e->getMessage()];} }
 public function index(Request $request) {
  Gate::authorize('viewAny',SynapseUser::class);
  foreach(['guests','admins','deactivated','locked'] as $key) if($request->query->has($key)) {
   $parsed=filter_var($request->query($key),FILTER_VALIDATE_BOOLEAN,FILTER_NULL_ON_FAILURE);
   if($parsed===null) abort(422,"فیلتر $key باید true یا false باشد.");
   $request->merge([$key=>$parsed]);
  }
  $filters=$request->validate(['from'=>['sometimes','regex:/^\d+$/'],'limit'=>['sometimes','integer','min:1','max:500'],'name'=>['sometimes','string','max:255'],'user_id'=>['sometimes','string','max:255'],'guests'=>['sometimes','boolean'],'admins'=>['sometimes','boolean'],'deactivated'=>['sometimes','boolean'],'locked'=>['sometimes','boolean'],'order_by'=>['sometimes','in:name,is_guest,admin,user_type,deactivated,shadow_banned,displayname,avatar_url,creation_ts,last_seen_ts'],'dir'=>['sometimes','in:f,b']]);
  $supportsLocked=$this->supportsLocked();if(!$supportsLocked)unset($filters['locked']);$filters['limit']??=50;$filters['from']??='0';$filters['guests']??=false;
  try{$result=$this->core->request('GET','/api/v1/users',$filters);$error=null;}catch(\Throwable $e){$result=['users'=>[],'total'=>0,'next_token'=>null];$error='دریافت فهرست کاربران از Synapse ناموفق بود: '.$e->getMessage();}
  return Inertia::render('Users/Index',['result'=>$result,'filters'=>$filters,'error'=>$error,'supportsLocked'=>$supportsLocked,'canCreate'=>Gate::allows('create',SynapseUser::class)]);
 }
 public function show(Request $request,string $userID) {
  Gate::authorize('view',[SynapseUser::class,$userID]);$id=$this->id($userID);$user=$this->core->request('GET',"/api/v1/users/$id");
  $devices=$this->optional('GET',"/api/v1/users/$id/devices");$rooms=$this->optional('GET',"/api/v1/users/$id/joined-rooms");$media=$this->optional('GET',"/api/v1/users/$id/media",['limit'=>100]);$mediaUsage=$this->optional('GET',"/api/v1/users/$id/media-usage");$ratelimit=$this->optional('GET',"/api/v1/users/$id/ratelimit");
  $canViewWhois=Gate::allows('viewWhois',SynapseUser::class);$whois=null;if($canViewWhois){$whois=$this->optional('GET',"/api/v1/users/$id/whois");if(!($whois['_unavailable']??false))$this->audit->record($request,'synapse.user.whois.view','matrix_user',$userID);}
  return Inertia::render('Users/Show',['user'=>$user,'devices'=>$devices,'rooms'=>$rooms,'whois'=>$whois,'media'=>$media,'mediaUsage'=>$mediaUsage,'ratelimit'=>$ratelimit,'supportsLocked'=>$this->supportsLocked(),'permissions'=>['update'=>Gate::allows('update',[SynapseUser::class,$userID]),'moderate'=>Gate::allows('moderate',[SynapseUser::class,$userID]),'delete'=>Gate::allows('delete',[SynapseUser::class,$userID]),'viewWhois'=>$canViewWhois]]);
 }
 public function store(Request $request) {
  Gate::authorize('create',SynapseUser::class);$data=$request->validate(['user_id'=>['required','string','regex:/^@[^\s]+:[^\s]+$/','max:255'],'password'=>['required','string','min:12','max:255'],'displayname'=>['nullable','string','max:255'],'admin'=>['sometimes','boolean'],'deactivated'=>['sometimes','boolean'],'locked'=>['sometimes','boolean']]);
  $userID=$data['user_id'];unset($data['user_id']);if(!$this->supportsLocked())unset($data['locked']);$this->core->request('PUT','/api/v1/users/'.$this->id($userID),[],$data);
  $this->audit->record($request,'synapse.user.create','matrix_user',$userID,['displayname'=>$data['displayname']??null,'admin'=>$data['admin']??false,'deactivated'=>$data['deactivated']??false,'locked'=>$data['locked']??false]);return redirect()->route('users.show',$userID)->with('success','کاربر ایجاد شد.');
 }
 public function update(Request $request,string $userID) {
  Gate::authorize('update',[SynapseUser::class,$userID]);$data=$request->validate(['displayname'=>['sometimes','nullable','string','max:255'],'admin'=>['sometimes','boolean'],'deactivated'=>['sometimes','boolean'],'locked'=>['sometimes','boolean']]);
  if(!$this->supportsLocked())unset($data['locked']);if(!$data)abort(422,'هیچ فیلدی برای به‌روزرسانی ارسال نشده است.');$this->core->request('PUT','/api/v1/users/'.$this->id($userID),[],$data);$this->audit->record($request,'synapse.user.update','matrix_user',$userID,$data);return back()->with('success','اطلاعات کاربر ذخیره شد.');
 }
 public function resetPassword(Request $request,string $userID) {
  Gate::authorize('update',[SynapseUser::class,$userID]);$data=$request->validate(['new_password'=>['required','string','min:12','max:255'],'logout_devices'=>['sometimes','boolean']]);$this->core->request('POST','/api/v1/users/'.$this->id($userID).'/reset-password',[],$data);$this->audit->record($request,'synapse.user.reset_password','matrix_user',$userID,['logout_devices'=>$data['logout_devices']??true]);return back()->with('success','گذرواژه بازنشانی شد.');
 }
 public function deactivate(Request $request,string $userID) {
  Gate::authorize('moderate',[SynapseUser::class,$userID]);$data=$request->validate(['erase'=>['sometimes','boolean']]);$this->core->request('POST','/api/v1/users/'.$this->id($userID).'/deactivate',[],$data);$this->audit->record($request,'synapse.user.deactivate','matrix_user',$userID,$data);return back()->with('success','حساب غیرفعال شد.');
 }
 public function suspend(Request $request,string $userID) {
  Gate::authorize('moderate',[SynapseUser::class,$userID]);$data=$request->validate(['suspend'=>['required','boolean']]);$this->core->request('PUT','/api/v1/users/'.$this->id($userID).'/suspend',[],$data);$this->audit->record($request,'synapse.user.suspend','matrix_user',$userID,$data);return back()->with('success','وضعیت تعلیق کاربر تغییر کرد.');
 }
 public function shadowBan(Request $request,string $userID) {
  Gate::authorize('moderate',[SynapseUser::class,$userID]);$ban=!$request->isMethod('delete')&&!$request->boolean('remove');$this->core->request($ban?'POST':'DELETE','/api/v1/users/'.$this->id($userID).'/shadow-ban');$this->audit->record($request,$ban?'synapse.user.shadow_ban':'synapse.user.shadow_unban','matrix_user',$userID);return back()->with('success',$ban?'Shadow Ban فعال شد.':'Shadow Ban برداشته شد.');
 }
 public function ratelimit(Request $request,string $userID) {
  Gate::authorize('update',[SynapseUser::class,$userID]);if($request->boolean('remove')){$this->core->request('DELETE','/api/v1/users/'.$this->id($userID).'/ratelimit');$payload=['removed'=>true];}else{$payload=$request->validate(['messages_per_second'=>['required','integer','min:0'],'burst_count'=>['required','integer','min:0']]);$this->core->request('POST','/api/v1/users/'.$this->id($userID).'/ratelimit',[],$payload);}
  $this->audit->record($request,'synapse.user.ratelimit','matrix_user',$userID,$payload);return back()->with('success','محدودیت نرخ به‌روزرسانی شد.');
 }
 public function deleteDevice(Request $request,string $userID,string $deviceID) {
  Gate::authorize('moderate',[SynapseUser::class,$userID]);$this->core->request('DELETE','/api/v1/users/'.$this->id($userID).'/devices/'.rawurlencode($deviceID));$this->audit->record($request,'synapse.user.device.delete','matrix_user',$userID,['device_id'=>$deviceID]);return back()->with('success','دستگاه حذف شد.');
 }
 public function deleteMedia(Request $request,string $userID) {
  Gate::authorize('delete',[SynapseUser::class,$userID]);$result=$this->core->request('DELETE','/api/v1/users/'.$this->id($userID).'/media',$request->only(['limit','order_by','dir']));$this->audit->record($request,'synapse.user.media.delete','matrix_user',$userID,['deleted_total'=>$result['deleted']??null,'possibly_more'=>$result['possibly_more']??false]);return back()->with('success','درخواست حذف رسانه ثبت شد.');
 }
}
