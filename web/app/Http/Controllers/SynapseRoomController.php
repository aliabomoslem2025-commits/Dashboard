<?php
namespace App\Http\Controllers;
use App\Services\AuditLogger;
use App\Services\CoreClient;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Gate;
use Inertia\Inertia;
class SynapseRoomController {
 public function __construct(private CoreClient $core,private AuditLogger $audit) {}
 private function id(string $id): string { return rawurlencode($id); }
 public function index(Request $r) { Gate::authorize('synapse-rooms-view');$query=$r->only(['from','limit','search_term','order_by','dir','public_rooms','empty_rooms']);try{$result=$this->core->request('GET','/api/v1/rooms',$query);$error=null;}catch(\Throwable $e){$result=['rooms'=>[],'total_rooms'=>0];$error=$e->getMessage();}return Inertia::render('Rooms/Index',['result'=>$result,'filters'=>$query,'error'=>$error]);}
 public function show(Request $r,string $roomID) { Gate::authorize('synapse-rooms-view');$id=$this->id($roomID);$room=$this->core->request('GET',"/api/v1/rooms/$id");$members=$this->core->request('GET',"/api/v1/rooms/$id/members");$state=$this->core->request('GET',"/api/v1/rooms/$id/state");$block=$this->core->request('GET',"/api/v1/rooms/$id/block");return Inertia::render('Rooms/Show',['room'=>$room,'members'=>$members,'state'=>$state,'block'=>$block]);}
 public function block(Request $r,string $roomID) { Gate::authorize('synapse-rooms-moderate');$data=$r->validate(['block'=>'required|boolean']);$this->core->request('PUT','/api/v1/rooms/'.$this->id($roomID).'/block',[],$data);$this->audit->record($r,'synapse.room.block','matrix_room',$roomID,$data);return back();}
 public function makeAdmin(Request $r,string $roomID) { Gate::authorize('synapse-rooms-moderate');$data=$r->validate(['user_id'=>'required|string|max:255']);$this->core->request('POST','/api/v1/rooms/'.$this->id($roomID).'/make-admin',[],$data);$this->audit->record($r,'synapse.room.make_admin','matrix_room',$roomID,$data);return back();}
 public function readOnly(Request $r,string $roomID) { Gate::authorize('synapse-rooms-moderate');$this->core->request('POST','/api/v1/rooms/'.$this->id($roomID).'/read-only',[],['enabled'=>true]);$this->audit->record($r,'synapse.room.read_only','matrix_room',$roomID,['enabled'=>true]);return back(); }
 public function delete(Request $r,string $roomID) { Gate::authorize('synapse-rooms-delete');$data=$r->validate(['purge'=>'required|boolean','force_purge'=>'sometimes|boolean','block'=>'sometimes|boolean','confirmation'=>'required|in:'.$roomID]);unset($data['confirmation']);$this->core->request('DELETE','/api/v1/rooms/'.$this->id($roomID),[],$data);$this->audit->record($r,'synapse.room.delete','matrix_room',$roomID,$data);return redirect()->route('rooms.index');}
}
