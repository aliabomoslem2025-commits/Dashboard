<?php
namespace App\Services;
use App\Models\AuditLog;
use Illuminate\Http\Request;
class AuditLogger { public function record(Request $request,string $action,?string $targetType=null,?string $targetId=null,array $payload=[]): AuditLog { return AuditLog::create(['user_id'=>$request->user()?->id,'action'=>$action,'target_type'=>$targetType,'target_id'=>$targetId,'payload'=>$payload,'ip_address'=>$request->ip(),'created_at'=>now()]); } }
