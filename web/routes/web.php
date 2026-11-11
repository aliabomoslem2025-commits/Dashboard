<?php
use Illuminate\\Support\\Facades\\Http;
use Illuminate\\Support\\Facades\\Route;
use Inertia\\Inertia;
Route::get('/', fn () => Inertia::render('Dashboard'))->name('dashboard');
Route::get('/health/core', function () { $response = Http::withToken(config('services.core.token'))->timeout(5)->get(rtrim(config('services.core.url'), '/').'/health'); return Inertia::render('CoreHealth', ['core' => $response->json(), 'reachable' => $response->successful()]); })->name('core.health');
