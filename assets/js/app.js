import { createApp } from 'vue'
import { RouterView } from 'vue-router'
import router from './router.js'
import emitter from './utils/mitt.ts'
import { useMittAxios } from './utils/axios'

createApp(RouterView).use(router).provide('axios', useMittAxios(emitter, {})).provide('mitt', emitter).mount('#app')
