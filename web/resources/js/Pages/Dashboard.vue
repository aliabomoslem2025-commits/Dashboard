<script setup lang="ts">
import { Head } from '@inertiajs/vue3'
const toggleTheme = () => { const theme = localStorage.getItem('matrix-theme') === 'dark' ? 'light' : 'dark'; localStorage.setItem('matrix-theme', theme); window.location.reload() }
</script>
<template><Head title="وضعیت سرور"/><v-app><v-app-bar color="surface" elevation="1"><v-app-bar-title>داشبورد مدیریت Matrix Synapse</v-app-bar-title><v-spacer/><v-btn icon="mdi-theme-light-dark" variant="text" @click="toggleTheme"/></v-app-bar><v-main><v-container class="py-8"><v-row><v-col cols="12" md="8"><v-card rounded="xl" class="pa-6"><v-card-title>نمای کلی سامانه</v-card-title><v-card-text>زیرساخت اولیهٔ پنل مدیریت Matrix آماده شده است. وضعیت سرویس مرکزی را بررسی کنید.</v-card-text><v-card-actions><v-btn color="primary" href="/health/core">بررسی mx-core</v-btn></v-card-actions></v-card></v-col></v-row></v-container></v-main></v-app></template>
