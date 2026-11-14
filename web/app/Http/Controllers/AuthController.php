<?php
namespace App\Http\Controllers;

use App\Models\User;
use Illuminate\Http\Request;
use Illuminate\Support\Facades\Auth;
use Illuminate\Support\Facades\Hash;
use Illuminate\Support\Facades\RateLimiter;
use Illuminate\Validation\ValidationException;
use Inertia\Inertia;

class AuthController
{
    public function create() { return Inertia::render('Auth/Login'); }

    public function store(Request $request)
    {
        $credentials = $request->validate(['username'=>['required','string','max:64'],'password'=>['required','string']]);
        $key = strtolower($credentials['username']).'|'.$request->ip();
        if (RateLimiter::tooManyAttempts($key, 5)) {
            throw ValidationException::withMessages(['username'=>'تعداد تلاش‌ها بیش از حد مجاز است. کمی بعد دوباره تلاش کنید.']);
        }
        $user = User::where('username', $credentials['username'])->first();
        if (!$user || !Hash::check($credentials['password'], $user->password)) {
            RateLimiter::hit($key, 60);
            throw ValidationException::withMessages(['username'=>'نام کاربری یا رمز عبور صحیح نیست.']);
        }
        if (!$user->is_active) {
            throw ValidationException::withMessages(['username'=>'این حساب غیرفعال است.']);
        }
        Auth::login($user, $request->boolean('remember'));
        RateLimiter::clear($key);
        $request->session()->regenerate();
        return redirect()->intended('/');
    }

    public function destroy(Request $request)
    {
        Auth::logout(); $request->session()->invalidate(); $request->session()->regenerateToken();
        return redirect('/login');
    }
}
