<script setup lang="ts">
import { Head, useForm } from '@inertiajs/vue3'
const form=useForm({email:'',password:'',remember:false})
const submit=()=>form.post('/login')
</script>
<template><Head title="ورود مدیران"/><div class="d-flex align-center justify-center" style="min-height:calc(100vh - 64px)"><v-card width="440" max-width="92vw" rounded="xl" class="pa-6"><v-card-title class="text-h5">ورود به پنل مدیریت</v-card-title><v-card-subtitle>دسترسی فقط برای کاربران ایجادشده توسط مدیر سیستم</v-card-subtitle><v-card-text><v-form @submit.prevent="submit"><v-text-field v-model="form.email" label="ایمیل" type="email" autocomplete="username" :error-messages="form.errors.email"/><v-text-field v-model="form.password" label="گذرواژه" type="password" autocomplete="current-password" :error-messages="form.errors.password"/><v-checkbox v-model="form.remember" label="مرا به خاطر بسپار" density="compact"/><v-btn type="submit" color="primary" block :loading="form.processing">ورود</v-btn></v-form></v-card-text></v-card></div></template>
