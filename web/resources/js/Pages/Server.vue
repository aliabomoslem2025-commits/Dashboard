<script setup lang="ts">
import { Head } from '@inertiajs/vue3'
defineProps<{version:{server_version?:string;python_version?:string};available:boolean;error:string|null}>()
</script>
<template><Head title="اطلاعات سرور"/><v-app><v-app-bar><v-app-bar-title>اطلاعات Synapse</v-app-bar-title></v-app-bar><v-main><v-container class="py-8"><v-row><v-col cols="12" md="6"><v-card rounded="xl" class="pa-5"><v-card-title>نسخه سرور</v-card-title><v-card-text><v-alert :type="available?'success':'error'" variant="tonal">{{ available ? `Synapse ${version.server_version || 'نامشخص'}` : error }}</v-alert><div class="mt-4">Python: {{ version.python_version || '—' }}</div></v-card-text></v-card></v-col></v-row></v-container></v-main></v-app></template>
