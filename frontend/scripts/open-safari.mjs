import { spawn } from 'node:child_process'

const url = process.argv[2] ?? 'http://localhost:5173'

const child = spawn('open', ['-a', 'Safari', url], {
  stdio: 'inherit',
})

child.on('exit', (code) => {
  process.exit(code ?? 0)
})
