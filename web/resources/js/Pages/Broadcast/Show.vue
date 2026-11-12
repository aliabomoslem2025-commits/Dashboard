<script setup lang="ts">
import { Head, router } from '@inertiajs/vue3'
import { onMounted,onUnmounted } from 'vue'
const props=defineProps<{broadcast:any;recipients:{data:any[];total:number};progress:{total:number;sent:number;failed:number}}>()
let timer:number|undefined
onMounted(()=>{timer=window.setInterval(()=>router.reload({only:['broadcast','recipients','progress'],preserveScroll:true,preserveState:true}),3000)})
onUnmounted(()=>{if(timer)window.clearInterval(timer)})
const sent=()=>Number(props.progress.sent||0)
const failed=()=>Number(props.progress.failed||0)
const total=()=>Number(props.progress.total||0)
const percent=()=>total()?Math.round((sent()+failed())/total()*100):0
const pending=()=>Math.max(0,total()-sent()-failed())
const statusColor=()=>props.broadcast.status==='completed'?'success':props.broadcast.status==='completed_with_errors'?'warning':props.broadcast.status==='failed'?'error':props.broadcast.status==='sending'?'info':'primary'
</script>
<template><Head :title="'ارسال #'+broadcast.id"/><v-container fluid class="pa-6 broadcast-detail"><v-btn href="/broadcasts" variant="text" prepend-icon="mdi-arrow-right">بازگشت</v-btn><div class="d-flex align-center justify-space-between flex-wrap ga-3 my-4"><div><h1 class="text-h4">وضعیت ارسال #{{ broadcast.id }}</h1><div class="text-medium-emphasis">به‌روزرسانی خودکار هر ۳ ثانیه</div></div><v-chip :color="statusColor()" variant="tonal">{{ broadcast.status }}</v-chip></div>
<v-card rounded="xl" class="pa-5 progress-card"><div class="text-body-1 mb-4" style="white-space:pre-wrap">{{ broadcast.message }}</div><v-progress-linear :model-value="percent()" height="13" rounded color="primary" class="progress-animated"/><div class="d-flex justify-space-between mt-3 text-caption text-medium-emphasis"><span>{{ percent() }}% تکمیل پردازش</span><span>{{ sent()+failed() }} از {{ total() }} پردازش شده</span></div><v-row class="mt-3"><v-col cols="4"><div class="status-stat"><div class="text-caption text-medium-emphasis">ارسال موفق</div><div class="text-h5 text-success">{{ sent() }}</div></div></v-col><v-col cols="4"><div class="status-stat"><div class="text-caption text-medium-emphasis">در انتظار</div><div class="text-h5">{{ pending() }}</div></div></v-col><v-col cols="4"><div class="status-stat"><div class="text-caption text-medium-emphasis">ناموفق</div><div class="text-h5 text-error">{{ failed() }}</div></div></v-col></v-row></v-card>
<v-card rounded="xl" class="pa-4 mt-5"><div class="d-flex align-center justify-space-between flex-wrap ga-2 mb-3"><v-card-title class="pa-0">وضعیت هر گیرنده</v-card-title><v-chip size="small" variant="tonal">{{ recipients.total }} گیرنده</v-chip></div><v-table><thead><tr><th>گیرنده</th><th>وضعیت</th><th>تلاش</th><th>شناسه رویداد</th><th>خطا</th></tr></thead><tbody><tr v-for="r in recipients.data" :key="r.id" class="recipient-row"><td>{{ r.user_id }}</td><td><v-chip size="small" :color="r.status==='sent'?'success':r.status==='failed'?'error':'warning'" variant="tonal">{{ r.status }}</v-chip></td><td>{{ r.attempts }}</td><td class="text-caption">{{ r.event_id||'—' }}</td><td class="text-error text-caption">{{ r.error||'—' }}</td></tr></tbody></v-table></v-card></v-container></template>
<style scoped>
.progress-card,.status-stat{transition:background-color .2s ease,transform .2s ease}
.status-stat{padding:10px 12px;border-radius:12px;background:rgba(var(--v-theme-on-surface),.035)}
.recipient-row{animation:row-in .25s ease both}
@keyframes row-in{from{opacity:0;transform:translateY(5px)}to{opacity:1;transform:translateY(0)}}
@media(prefers-reduced-motion:reduce){.recipient-row{animation:none}.progress-card,.status-stat{transition:none}}
</style>
