import { spawn } from 'node:child_process'

const url = process.argv[2] ?? 'http://localhost:5173'

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms))
}

async function isServerUp(targetUrl) {
  try {
    const controller = new AbortController()
    const t = setTimeout(() => controller.abort(), 800)
    const res = await fetch(targetUrl, { signal: controller.signal })
    clearTimeout(t)
    return res.ok
  } catch {
    return false
  }
}

async function waitForServer(targetUrl, timeoutMs) {
  const start = Date.now()
  while (Date.now() - start < timeoutMs) {
    if (await isServerUp(targetUrl)) return true
    await sleep(300)
  }
  return false
}

function openSafari(targetUrl) {
  return new Promise((resolve, reject) => {
    const child = spawn('open', ['-a', 'Safari', targetUrl], { stdio: 'inherit' })
    child.on('error', reject)
    child.on('exit', (code) => {
      if (code === 0 || code === null) resolve()
      else reject(new Error(`open Safari exited with code ${code}`))
    })
  })
}

const frontendCwd = new URL('../', import.meta.url).pathname

// If the dev server is already running, just open Safari.
if (await isServerUp(url)) {
  await openSafari(url)
  process.exit(0)
}

const vite = spawn('npm', ['run', 'dev'], {
  cwd: frontendCwd,
  stdio: ['inherit', 'pipe', 'pipe'],
  env: process.env,
})

vite.stdout.on('data', (chunk) => process.stdout.write(chunk))
vite.stderr.on('data', (chunk) => process.stderr.write(chunk))

let opened = false
const tryOpen = async () => {
  if (opened) return
  const ok = await waitForServer(url, 30_000)
  if (!ok) return
  opened = true
  try {
    await openSafari(url)
  } catch (err) {
    console.error(String(err))
  }
}

vite.stdout.on('data', () => {
  // Opportunistic: Vite logs appear; start checking immediately.
  void tryOpen()
})

vite.on('error', (err) => {
  console.error('Failed to start Vite:', err)
  process.exit(1)
})

vite.on('exit', (code) => {
  if (!opened) {
    console.error('Vite exited before the dev server became ready.')
  }
  process.exit(code ?? 0)
})

process.on('SIGINT', () => {
  vite.kill('SIGINT')
})
process.on('SIGTERM', () => {
  vite.kill('SIGTERM')
})

// Kick off readiness polling.
void tryOpen()
