<script setup lang="ts">
import { Transition } from 'vue'
defineProps<{title:string;value:string|number;icon?:string;caption?:string;to?:string}>()
</script>
<template><v-card rounded="xl" class="pa-4 stat-card" :href="to"><div class="d-flex align-center justify-space-between"><div class="text-medium-emphasis">{{ title }}</div><v-icon v-if="icon" :icon="icon" size="22"/></div><Transition name="count-pop" mode="out-in"><div :key="value" class="text-h4 font-weight-bold mt-3">{{ value }}</div></Transition><div v-if="caption" class="text-caption text-medium-emphasis mt-1">{{ caption }}</div></v-card></template>
<style scoped>
.stat-card{transition:transform .22s ease,box-shadow .22s ease}
.stat-card:hover{transform:translateY(-2px);box-shadow:0 8px 22px rgba(0,0,0,.10)}
.count-pop-enter-active,.count-pop-leave-active{transition:opacity .18s ease,transform .18s ease}
.count-pop-enter-from{opacity:0;transform:translateY(5px) scale(.97)}
.count-pop-leave-to{opacity:0;transform:translateY(-5px) scale(.97)}
@media(prefers-reduced-motion:reduce){.stat-card,.count-pop-enter-active,.count-pop-leave-active{transition:none}.stat-card:hover{transform:none}}
</style>
