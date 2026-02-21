import * as process from 'node:process'
import waitOn from 'wait-on'
import { spawn } from 'node:child_process'
import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const repoRoot = path.resolve(__dirname, '..')
const serviceDir = path.join(repoRoot, 'service')
const frontendDir = path.join(repoRoot, 'frontend')
const configSource = path.join(__dirname, 'configs', 'default', 'config.yaml')
const configDest = path.join(serviceDir, 'config.yaml')

export default function getRunner () {
  return new WackyTrackyTestRunner()
}

class WackyTrackyTestRunner {
  BASE_URL = 'http://localhost:8443/'
  serverProcess = null

  baseUrl () {
    return this.BASE_URL
  }

  async start () {
    console.log('      Building frontend...')
    const frontendBuild = spawn('npm', ['run', 'build'], {
      cwd: frontendDir,
      stdio: 'inherit',
      shell: true
    })
    await new Promise((resolve, reject) => {
      frontendBuild.on('close', (code) => (code === 0 ? resolve() : reject(new Error(`frontend build exited ${code}`))))
    })

    console.log('      Building service...')
    const serviceBuild = spawn('make', ['build'], {
      cwd: serviceDir,
      stdio: 'inherit'
    })
    await new Promise((resolve, reject) => {
      serviceBuild.on('close', (code) => (code === 0 ? resolve() : reject(new Error(`service build exited ${code}`))))
    })

    if (fs.existsSync(configSource)) {
      fs.copyFileSync(configSource, configDest)
    }

    const testdataDir = path.join(serviceDir, 'testdata', 'todotxt')
    fs.mkdirSync(testdataDir, { recursive: true })

    const binaryPath = path.join(serviceDir, 'wacky-tracky-server')
    if (!fs.existsSync(binaryPath)) {
      throw new Error('Service binary not found at ' + binaryPath)
    }

    console.log('      Starting WackyTracky server...')
    this.serverProcess = spawn(binaryPath, [], {
      cwd: serviceDir,
      env: { ...process.env }
    })

    this.serverProcess.stdout?.on('data', (d) => process.env.CI === 'true' && process.stdout.write(d))
    this.serverProcess.stderr?.on('data', (d) => process.env.CI === 'true' && process.stderr.write(d))
    this.serverProcess.on('close', (code) => {
      if (code != null) console.log('      Server exited with code', code)
    })

    await waitOn({ resources: [this.BASE_URL], timeout: 15000 })
    console.log('      Server started and reachable')
  }

  async stop () {
    if (!this.serverProcess) return
    if (this.serverProcess.exitCode != null) {
      console.log('      Server already exited')
      return
    }
    this.serverProcess.kill()
    console.log('      Server killed')
    await new Promise((r) => setTimeout(r, process.env.CI === 'true' ? 2000 : 300))
  }
}
