<script setup lang="ts">
defineProps<{title:string;value:string|number;icon?:string;caption?:string;to?:string}>()
</script>
<template><v-card rounded="xl" class="pa-4 stat-card" :href="to"><div class="d-flex align-center justify-space-between"><div class="text-medium-emphasis">{{ title }}</div><v-icon v-if="icon" :icon="icon" size="22"/></div><div class="text-h4 font-weight-bold mt-3">{{ value }}</div><div v-if="caption" class="text-caption text-medium-emphasis mt-1">{{ caption }}</div></v-card></template>
<style scoped>.stat-card{transition:transform .2s ease,box-shadow .2s ease}.stat-card:hover{transform:translateY(-2px)}@media(prefers-reduced-motion:reduce){.stat-card{transition:none}.stat-card:hover{transform:none}}</style>
