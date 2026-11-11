import { createApp, h } from 'vue'
import { createInertiaApp } from '@inertiajs/vue3'
import { resolvePageComponent } from 'laravel-vite-plugin/inertia-helpers'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'
import 'vuetify/styles'
import '@mdi/font/css/materialdesignicons.css'
const vuetify = createVuetify({ components, directives, theme: { defaultTheme: localStorage.getItem('matrix-theme') || 'light', themes: { light: { dark:false, colors:{ primary:'#3659a2', background:'#f5f6fa', surface:'#ffffff' } }, dark:{ dark:true, colors:{ primary:'#a9c3ff', background:'#111827', surface:'#1f2937' } } } } })
createInertiaApp({ resolve: name => resolvePageComponent(`./Pages/${name}.vue`, import.meta.glob('./Pages/**/*.vue')), setup({el,App,props,plugin}) { createApp({render:()=>h(App,props)}).use(plugin).use(vuetify).mount(el) } })
