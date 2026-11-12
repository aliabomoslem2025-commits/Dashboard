<script setup lang="ts">
import {Head,router,useForm} from '@inertiajs/vue3'
import {computed,ref} from 'vue'
type UserRow={name:string;displayname?:string|null;admin:boolean;deactivated:boolean;locked?:boolean;shadow_banned?:boolean;last_seen_ts?:number|null;creation_ts?:number}
const props=defineProps<{result:{users?:UserRow[];total?:number;next_token?:string|null};filters:Record<string,any>;error:string|null;supportsLocked:boolean;canCreate:boolean}>()
const search=ref(props.filters.name||props.filters.user_id||''),status=ref(String(props.filters.deactivated??'false')),locked=ref(String(props.filters.locked??'false')),admins=ref(props.filters.admins===undefined?'':String(props.filters.admins)),guests=ref(String(props.filters.guests??'false')),order=ref(props.filters.order_by||'name'),direction=ref(props.filters.dir||'f'),loading=ref(false),createOpen=ref(false)
const previousTokens=ref<string[]>([]),users=computed(()=>props.result.users||[])
const create=useForm({user_id:'',password:'',displayname:'',admin:false,deactivated:false,locked:false})
const statusLabel=(u:UserRow)=>u.deactivated?'غیرفعال':u.locked?'قفل‌شده':u.shadow_banned?'Shadow Ban':'فعال'
const filters=()=>({name:search.value||undefined,from:'0',limit:50,order_by:order.value,dir:direction.value,deactivated:status.value,locked:props.supportsLocked?locked.value:undefined,admins:admins.value===''?undefined:admins.value,guests:guests.value})
const apply=()=>{previousTokens.value=[];loading.value=true;router.get('/users',filters(),{preserveState:true,replace:true,onFinish:()=>loading.value=false})}
const nextPage=()=>{if(!props.result.next_token)return;previousTokens.value.push(String(props.filters.from??'0'));loading.value=true;router.get('/users',{...props.filters,from:props.result.next_token},{preserveState:true,onFinish:()=>loading.value=false})}
const previousPage=()=>{const from=previousTokens.value.pop();if(from===undefined)return;loading.value=true;router.get('/users',{...props.filters,from},{preserveState:true,onFinish:()=>loading.value=false})}
const createUser=()=>create.post('/users',{onSuccess:()=>{createOpen.value=false;create.reset()}})
</script>
<template><Head title="مدیریت کاربران"/><v-container fluid class="pa-4 pa-md-6">
 <div class="d-flex align-center justify-space-between flex-wrap ga-3 mb-5"><div><h1 class="text-h4">کاربران</h1><div class="text-medium-emphasis mt-1">مدیریت حساب‌های Matrix</div></div><v-btn v-if="canCreate" color="primary" prepend-icon="mdi-account-plus" @click="createOpen=true">ایجاد کاربر</v-btn></div>
 <v-card rounded="xl" class="pa-3 pa-md-5"><v-row dense align="center">
  <v-col cols="12" md="4"><v-text-field v-model="search" label="جستجو با شناسه یا نام نمایشی" prepend-inner-icon="mdi-magnify" clearable hide-details @keyup.enter="apply"/></v-col>
  <v-col cols="6" md="2"><v-select v-model="status" label="وضعیت حساب" :items="[{title:'فقط فعال‌ها',value:'false'},{title:'شامل غیرفعال‌ها',value:'true'}]" hide-details/></v-col>
  <v-col cols="6" md="2"><v-select v-model="admins" label="نوع حساب" :items="[{title:'همه',value:''},{title:'فقط مدیران',value:'true'},{title:'فقط کاربران عادی',value:'false'}]" hide-details/></v-col>
  <v-col cols="6" md="2"><v-select v-model="locked" label="حساب قفل‌شده" :items="[{title:'مخفی',value:'false'},{title:'شامل قفل‌شده‌ها',value:'true'}]" :disabled="!supportsLocked" hide-details/></v-col>
  <v-col cols="6" md="2"><v-select v-model="guests" label="مهمان‌ها" :items="[{title:'مخفی',value:'false'},{title:'شامل مهمان‌ها',value:'true'}]" hide-details/></v-col>
  <v-col cols="6" md="3"><v-select v-model="order" label="مرتب‌سازی" :items="[{title:'شناسه',value:'name'},{title:'نام نمایشی',value:'displayname'},{title:'آخرین فعالیت',value:'last_seen_ts'},{title:'زمان ایجاد',value:'creation_ts'}]" hide-details/></v-col>
  <v-col cols="6" md="2"><v-select v-model="direction" label="جهت مرتب‌سازی" :items="[{title:'صعودی',value:'f'},{title:'نزولی',value:'b'}]" hide-details/></v-col>
  <v-col cols="12" md="2"><v-btn block color="primary" @click="apply">اعمال فیلتر</v-btn></v-col>
 </v-row>
 <v-alert v-if="!supportsLocked" type="info" variant="tonal" class="mt-4">فیلتر و وضعیت قفل‌شده فقط از Synapse 1.93 به بعد پشتیبانی می‌شود.</v-alert><v-alert v-if="error" type="error" variant="tonal" class="mt-4">{{error}}</v-alert>
 <v-skeleton-loader v-if="loading" class="mt-4" type="table-row@8"/><template v-else>
  <div class="d-flex align-center justify-space-between mt-5 mb-2"><span class="text-body-2 text-medium-emphasis">کاربران یافت‌شده: {{result.total??users.length}}</span><v-chip size="small" variant="tonal">صفحه جاری</v-chip></div>
  <v-table class="user-table"><thead><tr><th>شناسه کاربر</th><th>نام نمایشی</th><th>وضعیت</th><th>مدیر</th><th>آخرین فعالیت</th><th class="text-end">عملیات</th></tr></thead>
   <TransitionGroup name="row" tag="tbody"><tr v-for="u in users" :key="u.name"><td class="font-weight-medium text-no-wrap">{{u.name}}</td><td>{{u.displayname||'—'}}</td><td><v-chip size="small" :color="u.deactivated?'error':u.locked?'warning':u.shadow_banned?'deep-orange':'success'" variant="tonal">{{statusLabel(u)}}</v-chip></td><td>{{u.admin?'بله':'خیر'}}</td><td>{{u.last_seen_ts?new Date(u.last_seen_ts).toLocaleString():'—'}}</td><td class="text-end"><v-btn size="small" variant="tonal" :href="'/users/'+encodeURIComponent(u.name)">جزئیات</v-btn></td></tr></TransitionGroup>
   <tbody v-if="!users.length"><tr><td colspan="6" class="text-center py-10 text-medium-emphasis">کاربری مطابق فیلترها یافت نشد.</td></tr></tbody>
  </v-table><div class="d-flex align-center justify-space-between flex-wrap ga-2 mt-4"><v-btn variant="tonal" prepend-icon="mdi-chevron-right" :disabled="previousTokens.length===0" @click="previousPage">صفحه قبل</v-btn><span class="text-caption text-medium-emphasis">next_token مستقیماً از Synapse دریافت می‌شود.</span><v-btn variant="tonal" append-icon="mdi-chevron-left" :disabled="!result.next_token" @click="nextPage">صفحه بعد</v-btn></div>
 </template></v-card>
 <v-dialog v-model="createOpen" max-width="600"><v-card rounded="xl"><v-card-title>ایجاد حساب Matrix</v-card-title><v-card-text><v-form @submit.prevent="createUser">
  <v-text-field v-model="create.user_id" label="شناسه کامل، مانند @user:example.org" :error-messages="create.errors.user_id"/><v-text-field v-model="create.password" label="گذرواژه اولیه (حداقل ۱۲ نویسه)" type="password" :error-messages="create.errors.password"/><v-text-field v-model="create.displayname" label="نام نمایشی"/><v-switch v-model="create.admin" label="مدیر سرور"/><v-switch v-model="create.deactivated" label="ایجاد به‌صورت غیرفعال"/><v-switch v-model="create.locked" label="قفل‌شده" :disabled="!supportsLocked"/><v-btn type="submit" color="primary" block :loading="create.processing" :disabled="create.password.length<12||!create.user_id.startsWith('@')">ایجاد کاربر</v-btn>
 </v-form></v-card-text><v-card-actions><v-spacer/><v-btn variant="text" @click="createOpen=false">بستن</v-btn></v-card-actions></v-card></v-dialog>
</v-container></template>
<style scoped>.row-enter-active,.row-leave-active{transition:opacity .22s ease,transform .22s ease}.row-enter-from{opacity:0;transform:translateY(7px)}@media(prefers-reduced-motion:reduce){.row-enter-active,.row-leave-active{transition:none}}@media(max-width:700px){.user-table{overflow-x:auto}.user-table :deep(table){min-width:760px}}</style>
