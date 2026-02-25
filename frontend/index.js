import './buildid.js'
import 'femtocrank/style.css';
import './style.css';

import { Bootloader } from './js/modules/Bootloader.js'
import { createApp } from 'vue'

import App from './resources/vue/App.vue'
import router from './js/router.js'

export function init() {
  const bootloader = new Bootloader()
  bootloader.init()

  const app = createApp(App)
  app.use(router)
  app.mount('#slot-app')
}
