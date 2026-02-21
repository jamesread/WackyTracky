import { createClient } from '@connectrpc/connect'
import { createConnectTransport } from '@connectrpc/connect-web'
import { WackyTrackyClientService } from '../gen/wacky-tracky/clientapi/v1/wt_pb.js'

/**
 * Sets up the Connect RPC client (window.client) and removes the boot overlay.
 * The Vue app uses window.client for all API calls.
 * When offline, the client still exists; API calls will fail and components handle that via useOffline.
 */
export class Bootloader {
  init() {
    this.setupClient()
    this.removeBootOverlay()
  }

  setupClient() {
    const baseUrl = '/api/'
    window.transport = createConnectTransport({ baseUrl })
    window.client = createClient(WackyTrackyClientService, window.transport)
  }

  removeBootOverlay() {
    const el = document.querySelector('#initMessages')
    if (el) el.remove()
  }
}
