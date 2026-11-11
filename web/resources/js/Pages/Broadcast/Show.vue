<script setup lang="ts">
import { Head, router } from '@inertiajs/vue3'
import { onMounted,onUnmounted } from 'vue'
const props=defineProps<{broadcast:any;recipients:{data:any[];total:number}}>()
let timer:number|undefined
onMounted(()=>{timer=window.setInterval(()=>router.reload({only:['broadcast','recipients']}),3000)})
onUnmounted(()=>{if(timer)window.clearInterval(timer)})
const sent=()=>props.recipients.data.filter(x=>x.status==='sent').length
const percent=()=>props.recipients.total?Math.round(sent()/props.recipients.total*100):0
</script>
<template><Head :title="'ارسال #'+broadcast.id"/><v-container fluid class="pa-6"><v-btn href="/broadcasts" variant="text" prepend-icon="mdi-arrow-right">بازگشت</v-btn><h1 class="text-h4 my-4">وضعیت ارسال #{{ broadcast.id }}</h1><v-card rounded="xl" class="pa-5"><div class="text-body-1 mb-3">{{ broadcast.message }}</div><v-progress-linear :model-value="percent()" height="12" rounded color="primary"/><div class="d-flex justify-space-between mt-3"><span>{{ broadcast.status }}</span><span>{{ sent() }} از {{ recipients.total }} ارسال شده</span></div><v-table class="mt-4"><thead><tr><th>گیرنده</th><th>وضعیت</th><th>تلاش</th><th>خطا</th></tr></thead><tbody><tr v-for="r in recipients.data" :key="r.id"><td>{{ r.user_id }}</td><td>{{ r.status }}</td><td>{{ r.attempts }}</td><td>{{ r.error||'—' }}</td></tr></tbody></v-table></v-card></v-container></template>
