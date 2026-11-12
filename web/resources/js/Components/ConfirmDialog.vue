<script setup lang="ts">
import { ref, watch } from 'vue'
const props=defineProps<{modelValue:boolean;title:string;message:string;confirmText?:string;requiredText?:string;danger?:boolean}>()
const emit=defineEmits<{ 'update:modelValue':[value:boolean]; confirm:[] }>()
const typed=ref('')
watch(()=>props.modelValue,()=>typed.value='')
const allowed=()=>!props.requiredText||typed.value===props.requiredText
</script>
<template><v-dialog :model-value="modelValue" max-width="520" @update:model-value="emit('update:modelValue',$event)"><v-card rounded="xl"><v-card-title>{{ title }}</v-card-title><v-card-text><p>{{ message }}</p><v-text-field v-if="requiredText" v-model="typed" class="mt-3" label="برای تأیید، عبارت خواسته‌شده را وارد کنید" :hint="requiredText" persistent-hint/></v-card-text><v-card-actions><v-spacer/><v-btn variant="text" @click="emit('update:modelValue',false)">انصراف</v-btn><v-btn :color="danger?'error':'primary'" :disabled="!allowed()" @click="emit('confirm');emit('update:modelValue',false)">{{ confirmText||'تأیید' }}</v-btn></v-card-actions></v-card></v-dialog></template>
