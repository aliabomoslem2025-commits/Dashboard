<script setup lang="ts">
import { computed } from 'vue'
import { router, usePage } from '@inertiajs/vue3'
import { useTheme } from 'vuetify'
const page=usePage()
const theme=useTheme()
const authenticated=computed(()=>Boolean((page.props as any).auth?.user))
const links=[{title:'داشبورد',href:'/'},{title:'کاربران',href:'/users'},{title:'اتاق‌ها',href:'/rooms'},{title:'مدیا و سهمیه‌ها',href:'/media'},{title:'فعالیت',href:'/activity'},{title:'مانیتورینگ',href:'/metrics'},{title:'ارسال پیام',href:'/broadcasts'},{title:'گزارش رویدادها',href:'/audit-logs'}]
const toggleTheme=()=>{const next=theme.global.name.value==='dark'?'light':'dark';theme.global.name.value=next;localStorage.setItem('matrix-theme',next)}
const logout=()=>router.post('/logout')
</script>
<template>
 <v-app>
  <v-app-bar color="surface" elevation="1">
   <v-app-bar-title>مدیریت Matrix Synapse</v-app-bar-title>
   <v-menu>
    <template #activator="{props}"><v-btn v-bind="props" icon="mdi-menu" variant="text" aria-label="منوی صفحات"/></template>
    <v-list density="compact"><v-list-item v-for="link in links" :key="link.href" :title="link.title" :href="link.href"/></v-list>
   </v-menu>
   <v-btn icon="mdi-theme-light-dark" variant="text" aria-label="تغییر پوسته" @click="toggleTheme"/>
   <v-btn v-if="authenticated" variant="text" @click="logout">خروج</v-btn>
  </v-app-bar>
  <v-main><slot/></v-main>
 </v-app>
</template>
