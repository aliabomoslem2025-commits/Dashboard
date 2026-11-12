<script setup lang="ts">
import { Head } from '@inertiajs/vue3'
defineProps<{core:{ok?:boolean;data?:{status?:string;service?:string}}|null;reachable:boolean}>()
</script>
<template><Head title="وضعیت mx-core"/><v-container class="py-8"><v-card rounded="xl" class="pa-6" max-width="760"><v-card-title>اتصال Go Core</v-card-title><v-card-text><v-alert :type="reachable && core?.ok ? 'success' : 'error'" variant="tonal">{{ reachable && core?.ok ? 'سرویس mx-core در دسترس است.' : 'اتصال به mx-core برقرار نشد.' }}</v-alert></v-card-text><v-card-actions><v-btn href="/" variant="text">بازگشت</v-btn></v-card-actions></v-card></v-container></template>
