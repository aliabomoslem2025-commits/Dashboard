<?php
use App\Http\Controllers\AuthController;
use App\Http\Controllers\ServerController;
use Illuminate\Support\Facades\Http;
use Illuminate\Support\Facades\Route;
use Inertia\Inertia;
Route::middleware('guest')->group(function(){Route::get('/login',[AuthController::class,'create'])->name('login');Route::post('/login',[AuthController::class,'store'])->name('login.store');});
Route::post('/logout',[AuthController::class,'destroy'])->middleware('auth')->name('logout');
Route::middleware(['auth'])->group(function(){Route::get('/',fn()=>Inertia::render('Dashboard'))->name('dashboard');Route::get('/health/core',function(){$response=Http::withToken(config('services.core.token'))->timeout(5)->get(rtrim(config('services.core.url'),'/').'/health');return Inertia::render('CoreHealth',['core'=>$response->json(),'reachable'=>$response->successful()]);})->name('core.health');Route::get('/server',[ServerController::class,'show'])->name('server.show');});
