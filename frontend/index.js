import './buildid.js'
import 'femtocrank/style.css';
import './style.css';

import { purgeLegacyServiceWorkerCaches } from './js/modules/serviceWorkerMigration.js'
import { Bootloader } from './js/modules/Bootloader.js'
import { createApp } from 'vue'

import App from './resources/vue/App.vue'
import router from './js/router.js'

export function init() {
  purgeLegacyServiceWorkerCaches()

  const bootloader = new Bootloader()
  bootloader.init()

  const app = createApp(App)
  app.use(router)
  app.mount('#slot-app')
}
