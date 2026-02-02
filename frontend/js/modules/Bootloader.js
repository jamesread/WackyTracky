import { DBAL } from './DBAL.js'

import { setBootMessage } from '../firmware/util.js'
import { createClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'

import { WackyTrackyClientService } from './../gen/wacky-tracky/clientapi/v1/wt_pb.js'

/**
The bootloader sets up essential browser features, imports all the needed
"core" functionality, and gets the app to a point where it should be able to
render either a working UI, or sensible error messages.

The bootloader hands off to the UiManager when complete.
*/
export class Bootloader {
  init () {
    setBootMessage('Bootloader init')

    setBootMessage('Waiting for database')

    this.setupClient();

    window.dbal = new DBAL()
    window.dbal.open(() => { this.initAfterDatabase() })
  }

  setupClient() {
    let baseUrl = '/api/'

    window.transport = createConnectTransport({
      baseUrl: baseUrl,
    })

    window.client = createClient(WackyTrackyClientService, window.transport);
  }

  async initAfterDatabase () {
    setBootMessage('Database is OK')

    let res = await window.client.init();

    this.initSuccess(res)
  }

  /**
  At this stage, the critical "boot" path is complete.
  We have the minimum browser features, core javascript, etc.
  */
  initSuccess (res) {
    window.uimanager.onBootloaderSuccess(res)
  }

  initFailure (a, b, c) {
    if (a != null && a.toString().includes('Failed to fetch')) {
      // setBootMessage("Failed to fetch during init, are you offline?")
      window.uimanager.onBootloaderOffline()
    } else {
      setBootMessage('Unknown init failure.')
    }
  }
}
