<?php
namespace App\Http\Controllers;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\RateLimiter;
use Illuminate\Validation\ValidationException;
use Inertia\Inertia;
class AuthController { public function create(){return Inertia::render('Auth/Login');} public function store(Request $request){$credentials=$request->validate(['email'=>['required','email'],'password'=>['required','string']]);$key=strtolower($credentials['email']).'|'.$request->ip();if(RateLimiter::tooManyAttempts($key,5)){throw ValidationException::withMessages(['email'=>'تعداد تلاش‌ها بیش از حد مجاز است. کمی بعد دوباره تلاش کنید.']);}if(!Auth::attempt($credentials,$request->boolean('remember'))){RateLimiter::hit($key,60);throw ValidationException::withMessages(['email'=>'اطلاعات ورود صحیح نیست.']);}if(!Auth::user()->is_active){Auth::logout();throw ValidationException::withMessages(['email'=>'این حساب غیرفعال است.']);}RateLimiter::clear($key);$request->session()->regenerate();return redirect()->intended('/');}public function destroy(Request $request){Auth::logout();$request->session()->invalidate();$request->session()->regenerateToken();return redirect('/login');} }
