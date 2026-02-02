import './buildid.js'
import 'femtocrank/style.css';

import { UiManager } from './js/modules/UiManager.js'
import { Bootloader } from './js/modules/Bootloader.js'
import { createApp } from 'vue'

import App from './resources/vue/App.vue'
import router from './js/router.js'

export function init() {
  window.uimanager = new UiManager()

  window.bootloader = new Bootloader()
  window.bootloader.init()

  const app = createApp(App)
  app.use(router)
  app.mount('#slot-app')
}
