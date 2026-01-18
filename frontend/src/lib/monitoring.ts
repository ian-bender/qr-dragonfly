import type { App } from 'vue'
import type { Router } from 'vue-router'

interface SentryConfig {
  dsn: string
  environment: string
  release?: string
  tracesSampleRate?: number
}

// Sentry is optional - only import if configured
let Sentry: any = null

export function initSentry(app: App, router: Router, config: SentryConfig) {
  if (!config.dsn) {
    console.log('Sentry DSN not configured, error tracking disabled')
    return
  }

  // Dynamically import Sentry only if DSN is configured
  // @ts-ignore - Sentry is an optional dependency
  import('@sentry/vue')
    .then((SentryModule) => {
      Sentry = SentryModule
      
      Sentry.init({
        app,
        dsn: config.dsn,
        environment: config.environment,
        release: config.release,
        integrations: [
          Sentry.browserTracingIntegration({ router }),
          Sentry.replayIntegration({
            maskAllText: true,
            blockAllMedia: true,
          }),
        ],
        
        // Performance Monitoring
        tracesSampleRate: config.tracesSampleRate || 0.1, // 10% of transactions
        
        // Session Replay
        replaysSessionSampleRate: 0.1, // 10% of sessions
        replaysOnErrorSampleRate: 1.0, // 100% of sessions with errors
        
        // Filter out expected errors
        beforeSend(event: any, hint: any) {
          // Don't send 401/403 errors (expected authentication failures)
          if (event.exception?.values?.[0]?.value?.includes('401') || 
              event.exception?.values?.[0]?.value?.includes('403')) {
            return null
          }
          
          // Don't send network errors for offline users
          if (hint.originalException instanceof Error) {
            if (hint.originalException.message.includes('Network Error') && !navigator.onLine) {
              return null
            }
          }
          
          return event
        },
      })

      console.log(`Sentry initialized: ${config.environment}`)
    })
    .catch((err) => {
      console.warn('Failed to load Sentry:', err)
    })
}

export function captureError(error: Error, context?: Record<string, any>) {
  if (!Sentry) {
    console.error('Sentry not initialized:', error)
    return
  }
  
  Sentry.withScope((scope: any) => {
    if (context) {
      Object.entries(context).forEach(([key, value]) => {
        scope.setExtra(key, value)
      })
    }
    Sentry.captureException(error)
  })
}

export function captureMessage(message: string, level: 'info' | 'warning' | 'error' = 'info') {
  if (!Sentry) {
    console.warn('Sentry not initialized:', message)
    return
  }
  
  Sentry.captureMessage(message, level)
}

export function setUserContext(user: { id: string; email?: string; userType?: string }) {
  if (!Sentry) return
  
  Sentry.setUser({
    id: user.id,
    email: user.email,
    userType: user.userType,
  })
}

export function clearUserContext() {
  if (!Sentry) return
  
  Sentry.setUser(null)
}
