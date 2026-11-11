<?php
namespace App\Http\Controllers;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Gate;
use Inertia\Inertia;
class AuditLogController {
 public function index(Request $r){Gate::authorize('view-audit-logs');$logs=DB::table('audit_logs')->when($r->filled('action'),fn($q)=>$q->where('action','ilike','%'.$r->string('action').'%'))->when($r->filled('target'),fn($q)=>$q->where('target_id','ilike','%'.$r->string('target').'%'))->orderByDesc('id')->paginate(50)->withQueryString();return Inertia::render('Audit/Index',['logs'=>$logs,'filters'=>$r->only(['action','target'])]);}
}
