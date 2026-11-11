<?php
namespace App\Http\Controllers;
use App\Services\AuditLogger;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\DB;
use Illuminate\Support\Facades\Gate;
use Inertia\Inertia;
class MetricsController {
 public function __construct(private AuditLogger $audit) {}
 public function stream(Request $request){Gate::authorize('view-dashboard');return response()->stream(function() use($request){try{$client=new \\GuzzleHttp\\Client(['timeout'=>0,'connect_timeout'=>5]);$upstream=$client->request('GET',rtrim(config('services.core.url'),'/').'/api/v1/metrics/stream',['headers'=>['Authorization'=>'Bearer '.config('services.core.token'),'Accept'=>'text/event-stream'],'stream'=>true,'http_errors'=>false]);if($upstream->getStatusCode()!==200){echo "event: error\\ndata: {\\\"message\\\":\\\"metrics upstream unavailable\\\"}\\n\\n";return;}$body=$upstream->getBody();while(!$body->eof()&&!connection_aborted()){$chunk=$body->read(4096);if($chunk!==''){echo $chunk;if(ob_get_level()>0){@ob_flush();}flush();}}}catch(\\Throwable $e){echo "event: error\\ndata: {\\\"message\\\":\\\"metrics stream interrupted\\\"}\\n\\n";}},200,['Content-Type'=>'text/event-stream','Cache-Control'=>'no-cache, no-transform','Connection'=>'keep-alive','X-Accel-Buffering'=>'no']);}
 public function index(){Gate::authorize('view-dashboard');$latest=DB::table('system_metrics')->orderByDesc('sampled_at')->first();$history=DB::table('system_metrics')->orderByDesc('sampled_at')->limit(120)->get()->reverse()->values();$thresholds=DB::table('metric_thresholds')->pluck('value','name');return Inertia::render('Metrics/Index',['latest'=>$latest,'history'=>$history,'thresholds'=>$thresholds]);}
 public function updateThresholds(Request $r){Gate::authorize('manage-server');$data=$r->validate(['cpu_percent'=>'required|numeric|min:1|max:100','ram_percent'=>'required|numeric|min:1|max:100','disk_percent'=>'required|numeric|min:1|max:100']);foreach($data as $name=>$value)DB::table('metric_thresholds')->updateOrInsert(['name'=>$name],['value'=>$value,'updated_at'=>now(),'created_at'=>now()]);$this->audit->record($r,'metrics.thresholds.update','metric_thresholds','global',$data);return back()->with('success','آستانه‌های هشدار ذخیره شد.');}
}
