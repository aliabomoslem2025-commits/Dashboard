<script setup lang="ts">
import { Head,router } from '@inertiajs/vue3'
import { ref } from 'vue'
const props=defineProps<{logs:{data:any[];total:number};filters:any}>()
const action=ref(props.filters.action||''),target=ref(props.filters.target||'')
const apply=()=>router.get('/audit-logs',{action:action.value,target:target.value},{preserveState:true})
</script>
<template><Head title="گزارش رویدادها"/><v-container fluid class="pa-6"><h1 class="text-h4 mb-5">گزارش رویدادها (Audit Log)</h1><v-card rounded="xl" class="pa-4"><v-row><v-col cols="12" md="5"><v-text-field v-model="action" label="فیلتر عملیات"/></v-col><v-col cols="12" md="5"><v-text-field v-model="target" label="شناسه هدف"/></v-col><v-col cols="12" md="2"><v-btn block color="primary" @click="apply">فیلتر</v-btn></v-col></v-row><v-table><thead><tr><th>زمان</th><th>کاربر مدیر</th><th>عملیات</th><th>هدف</th><th>جزئیات</th><th>IP</th></tr></thead><tbody><tr v-for="log in logs.data" :key="log.id"><td>{{ log.created_at }}</td><td>{{ log.user_id??'سیستم' }}</td><td><v-chip size="small">{{ log.action }}</v-chip></td><td>{{ log.target_type }} / {{ log.target_id }}</td><td><pre class="text-caption">{{ JSON.stringify(log.payload||{},null,2) }}</pre></td><td>{{ log.ip_address||'—' }}</td></tr><tr v-if="!logs.data.length"><td colspan="6" class="text-center pa-6">رویدادی ثبت نشده است.</td></tr></tbody></v-table><div class="text-caption mt-3">تعداد رویدادها: {{ logs.total }}</div></v-card></v-container></template>
