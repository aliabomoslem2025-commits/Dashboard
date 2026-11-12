<script setup lang="ts">
import { Head, useForm } from '@inertiajs/vue3'
import { onMounted,onUnmounted,ref } from 'vue'
const props=defineProps<{latest:any;history:any[];thresholds:Record<string,number>}>()
const latest=ref(props.latest),history=ref([...props.history])
const form=useForm({cpu_percent:Number(props.thresholds.cpu_percent||85),ram_percent:Number(props.thresholds.ram_percent||85),disk_percent:Number(props.thresholds.disk_percent||90)})
const percent=(a:number,b:number)=>b?Math.min(100,a/b*100):0
const fmtBytes=(n:number)=>{if(!Number.isFinite(n)||n<0)return '—';if(n===0)return '0 B';const u=['B','KB','MB','GB','TB'];const i=Math.min(Math.floor(Math.log(n)/Math.log(1024)),u.length-1);return (n/Math.pow(1024,i)).toFixed(i?1:0)+' '+u[i]}
let stream:EventSource|undefined
onMounted(()=>{stream=new EventSource('/metrics/stream');stream.addEventListener('metrics',(event)=>{try{const data=JSON.parse((event as MessageEvent).data);if(data){latest.value=data;history.value=[...history.value,data].slice(-120)}}catch{}})})
onUnmounted(()=>stream?.close())
const trend=()=>history.value.map((x,i)=>{const max=Math.max(1,...history.value.map(y=>Number(y.cpu_percent||0)));return (i*100/(history.value.length-1||1))+','+(100-Number(x.cpu_percent||0)*100/max)}).join(' ')
const alerts=()=>[{label:'CPU',value:Number(latest.value?.cpu_percent||0),limit:form.cpu_percent},{label:'RAM',value:percent(latest.value?.ram_used_bytes||0,latest.value?.ram_total_bytes||0),limit:form.ram_percent},{label:'Disk',value:percent(latest.value?.disk_used_bytes||0,latest.value?.disk_total_bytes||0),limit:form.disk_percent}].filter(x=>x.value>=x.limit)
const prometheus=()=>{const raw=latest.value?.prometheus;if(!raw)return [];if(typeof raw==='string'){try{return Object.entries(JSON.parse(raw)).slice(0,30)}catch{return []}}return Object.entries(raw).filter(([k])=>k!=='error').slice(0,30)}
</script>
<template><Head title="مانیتورینگ سرور"/><v-container fluid class="pa-6 metrics-page"><div class="d-flex justify-space-between align-center flex-wrap ga-3 mb-5"><div><h1 class="text-h4">مانیتورینگ زنده</h1><div class="text-medium-emphasis">آخرین نمونه: {{ latest?.sampled_at||'در انتظار اولین نمونه‌برداری' }}</div></div><v-chip color="success" variant="tonal" prepend-icon="mdi-access-point">نمونه‌برداری هر ۱۵ ثانیه</v-chip></div>
<v-skeleton-loader v-if="!latest" type="card,card,card" class="mb-4"/>
<template v-else>
<v-alert v-for="a in alerts()" :key="a.label" type="warning" variant="tonal" class="mb-2">هشدار {{ a.label }}: {{ a.value.toFixed(1) }}% (آستانه {{ a.limit }}%)</v-alert>
<v-row><v-col v-for="item in [{name:'CPU',value:Number(latest.cpu_percent||0),detail:'پردازنده'},{name:'RAM',value:percent(latest.ram_used_bytes||0,latest.ram_total_bytes||0),detail:fmtBytes(latest.ram_used_bytes)+' / '+fmtBytes(latest.ram_total_bytes)},{name:'Disk',value:percent(latest.disk_used_bytes||0,latest.disk_total_bytes||0),detail:fmtBytes(latest.disk_used_bytes)+' / '+fmtBytes(latest.disk_total_bytes)}]" :key="item.name" cols="12" md="4"><v-card rounded="xl" class="pa-5 text-center metric-card"><v-progress-circular :model-value="item.value" :size="130" :width="12" color="primary">{{ item.value.toFixed(1) }}%</v-progress-circular><div class="text-h6 mt-3">{{ item.name }}</div><div class="text-caption text-medium-emphasis">{{ item.detail }}</div></v-card></v-col></v-row>
<v-row class="mt-2"><v-col v-for="item in [{label:'کل کاربران',value:latest.users_total},{label:'اتاق‌ها',value:latest.rooms_total},{label:'آنلاین در ۵ دقیقه',value:latest.online_users},{label:'اندازه دیتابیس',value:fmtBytes(Number(latest.database_bytes||0))}]" :key="item.label" cols="6" md="3"><v-card rounded="xl" class="pa-4 metric-card"><div class="text-medium-emphasis">{{ item.label }}</div><div class="text-h5 font-weight-bold mt-2">{{ item.value??'—' }}</div></v-card></v-col></v-row>
<v-card rounded="xl" class="pa-5 mt-5"><v-card-title>روند CPU · آخرین {{ history.length }} نمونه</v-card-title><svg viewBox="0 0 100 100" preserveAspectRatio="none" width="100%" height="150"><polyline :points="trend()" fill="none" stroke="currentColor" stroke-width="1.5" vector-effect="non-scaling-stroke"/></svg></v-card>
<v-card rounded="xl" class="pa-5 mt-5"><v-card-title>متریک‌های Prometheus Synapse</v-card-title><v-table v-if="prometheus().length" density="compact"><thead><tr><th>متریک</th><th class="text-end">مقدار</th></tr></thead><tbody><tr v-for="[key,value] in prometheus()" :key="key"><td class="text-break">{{ key }}</td><td class="text-end">{{ Number(value).toLocaleString() }}</td></tr></tbody></v-table><div v-else class="text-medium-emphasis pa-3">متریک Prometheus دریافت نشده است. مقدار SYNAPSE_PROMETHEUS_URL را بررسی کنید.</div></v-card>
</template>
<v-card rounded="xl" class="pa-5 mt-5"><v-card-title>آستانه‌های هشدار منابع</v-card-title><v-form @submit.prevent="form.put('/metrics/thresholds')"><v-row><v-col cols="12" md="4"><v-text-field v-model.number="form.cpu_percent" label="CPU درصد" type="number" min="1" max="100" suffix="%"/></v-col><v-col cols="12" md="4"><v-text-field v-model.number="form.ram_percent" label="RAM درصد" type="number" min="1" max="100" suffix="%"/></v-col><v-col cols="12" md="4"><v-text-field v-model.number="form.disk_percent" label="Disk درصد" type="number" min="1" max="100" suffix="%"/></v-col></v-row><v-btn type="submit" color="primary" :loading="form.processing">ذخیره آستانه‌ها</v-btn></v-form></v-card></v-container></template>
<style scoped>
.metric-card{transition:transform .22s ease,box-shadow .22s ease}
.metric-card:hover{transform:translateY(-3px);box-shadow:0 10px 26px rgba(0,0,0,.10)}
@media(prefers-reduced-motion:reduce){.metric-card{transition:none}.metric-card:hover{transform:none}}
</style>
