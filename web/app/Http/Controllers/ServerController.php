<?php
namespace App\Http\Controllers;
use App\Services\CoreClient;
use Illuminate\Support\Facades\Gate;
use Inertia\Inertia;
class ServerController { public function show(CoreClient $core){Gate::authorize('view-dashboard');try{$version=$core->getServerVersion();$available=true;$error=null;}catch(\Throwable $e){$version=[];$available=false;$error='ارتباط با Synapse برقرار نشد.';}return Inertia::render('Server',['version'=>$version,'available'=>$available,'error'=>$error]);} }
