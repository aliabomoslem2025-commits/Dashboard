 <script setup lang="ts">
import { Head, useForm } from '@inertiajs/vue3'
defineOptions({ layout: false })
const form = useForm({ username: '', password: '', remember: false })
const submit = () => form.post('/login')
</script>

<template>
  <Head title="ورود به پنل مدیریت ماتریکس" />
  <v-app class="login-app">
    <main class="login-page" dir="rtl">
      <section class="login-panel" aria-label="فرم ورود">
        <div class="login-card">
          <header class="brand">
            <div class="brand-copy"><div class="brand-title">پنل مدیریت ماتریکس</div><div class="brand-server">[YOUR SERVER]</div></div>
            <div class="brand-mark" aria-hidden="true"><span>م</span></div>
          </header>
          <div class="form-heading"><h1>خوش آمدید</h1><p>با حساب مدیر پنل وارد شوید.</p></div>
          <v-form class="login-form" @submit.prevent="submit">
            <div class="field-group">
              <label for="login-username">نام کاربری مدیر</label>
              <v-text-field id="login-username" v-model="form.username" class="login-input" placeholder="admin" autocomplete="username" variant="outlined" density="comfortable" hide-details="auto" :error-messages="form.errors.username" required />
            </div>
            <div class="field-group">
              <label for="login-password">رمز عبور</label>
              <v-text-field id="login-password" v-model="form.password" class="login-input" placeholder="••••••••" type="password" autocomplete="current-password" variant="outlined" density="comfortable" hide-details="auto" :error-messages="form.errors.password" required />
            </div>
            <v-checkbox v-model="form.remember" class="remember-checkbox" label="مرا به خاطر بسپار" color="primary" density="compact" hide-details />
            <v-btn class="login-button" color="primary" type="submit" block size="large" :loading="form.processing" :disabled="form.processing">ورود</v-btn>
          </v-form>
          <p class="login-footnote">دسترسی فقط برای مدیران سرور مجاز است.</p>
        </div>
      </section>
      <section class="intro-panel" aria-label="امکانات پنل">
        <div class="intro-content">
          <h2>کنترل کامل سرور Synapse<br />در یک پنل</h2>
          <div class="feature-grid">
            <article class="feature-card"><v-icon icon="mdi-account-group-outline" size="23" /><div><h3>کاربران</h3><p>سهمیه آپلود و محدودیت‌ها</p></div></article>
            <article class="feature-card"><v-icon icon="mdi-forum-outline" size="23" /><div><h3>اتاق‌ها</h3><p>مدیریت و کنترل دسترسی</p></div></article>
            <article class="feature-card"><v-icon icon="mdi-chart-box-outline" size="23" /><div><h3>پایش منابع</h3><p>مصرف و فعالیت کاربران</p></div></article>
            <article class="feature-card"><v-icon icon="mdi-bullhorn-outline" size="23" /><div><h3>پیام همگانی</h3><p>ارسال به همه یا یک کاربر</p></div></article>
          </div>
        </div>
      </section>
    </main>
  </v-app>
</template>

<style>
.login-app{min-height:100vh;background:#f4f5fa;font-family:inherit}.login-page{display:grid;grid-template-columns:46% 54%;min-height:100vh;direction:ltr;background:#f4f5fa}.login-panel{display:flex;align-items:center;justify-content:center;min-width:0;padding:48px;direction:rtl;background:#f4f5fa}.login-card{width:min(100%,492px);padding:44px 36px 40px;border-radius:21px;background:#fff;box-shadow:0 12px 38px rgba(29,38,91,.10)}.brand{display:flex;align-items:center;justify-content:flex-start;gap:14px;margin-bottom:34px}.brand-copy{min-width:0}.brand-title{color:#202331;font-size:21px;font-weight:750;line-height:1.5}.brand-server{margin-top:2px;color:#626b83;direction:ltr;text-align:right;font-size:11px;letter-spacing:.04em}.brand-mark{display:grid;width:44px;height:44px;flex:0 0 44px;place-items:center;border-radius:13px;background:#394bb2;color:white;font-size:27px;font-weight:500}.form-heading{margin-bottom:26px}.form-heading h1{margin:0 0 18px;color:#202331;font-size:27px;font-weight:800}.form-heading p{margin:0;color:#687087;font-size:13px}.login-form{display:flex;flex-direction:column;gap:17px}.field-group{display:flex;flex-direction:column;gap:8px}.field-group label{color:#282d3b;font-size:12px;font-weight:700}.login-input .v-field{border-radius:12px;background:#fff}.login-input .v-field__outline{color:#c7ccda}.login-input input{font-size:14px}.remember-checkbox{align-self:flex-start;margin-top:-2px;color:#687087}.login-button{min-height:52px!important;border-radius:12px;background:#394bb2!important;box-shadow:0 5px 12px rgba(57,75,178,.27);color:#fff!important;font-size:15px;font-weight:750;letter-spacing:0}.login-footnote{margin:22px 0 0;color:#687087;text-align:center;font-size:11px}.intro-panel{display:flex;align-items:center;justify-content:center;min-width:0;padding:64px clamp(42px,8.5vw,116px);direction:rtl;background:#1c267f;color:#fff}.intro-content{width:100%;max-width:650px}.intro-content h2{margin:0 0 34px;color:#fff;font-size:clamp(27px,2.8vw,36px);font-weight:800;line-height:1.65}.feature-grid{display:grid;grid-template-columns:repeat(2,minmax(0,1fr));gap:16px}.feature-card{display:flex;align-items:flex-start;gap:13px;min-height:90px;padding:20px 18px;border-radius:17px;background:#29388f;color:#fff}.feature-card .v-icon{margin-top:1px;color:#dce3ff}.feature-card h3{margin:0 0 7px;color:#fff;font-size:14px;font-weight:750}.feature-card p{margin:0;color:#c2cbed;font-size:11px;line-height:1.8}@media(max-width:960px){.login-page{grid-template-columns:1fr}.login-panel{grid-row:1;padding:42px 24px}.intro-panel{grid-row:2;padding:42px 28px}.intro-content{max-width:600px}.intro-content h2{font-size:27px}}@media(max-width:520px){.login-panel{padding:22px 14px}.login-card{padding:30px 22px 28px;border-radius:17px}.brand{margin-bottom:27px}.brand-title{font-size:18px}.form-heading h1{font-size:24px}.intro-panel{padding:34px 18px}.intro-content h2{margin-bottom:24px;font-size:24px}.feature-grid{gap:10px}.feature-card{min-height:86px;gap:8px;padding:15px 11px}.feature-card h3{font-size:12px}.feature-card p{font-size:10px}}@media(prefers-reduced-motion:reduce){*,*::before,*::after{scroll-behavior:auto!important;transition-duration:.01ms!important;animation-duration:.01ms!important;animation-iteration-count:1!important}}
</style>
