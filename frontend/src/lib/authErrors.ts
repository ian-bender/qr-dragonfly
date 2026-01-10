import { ApiError } from '../api'

export function authErrorMessage(err: unknown): string {
  if (err instanceof ApiError) {
    const payload = err.payload as any
    const code = payload?.error
    if (typeof code === 'string' && code.trim()) {
      switch (code) {
        case 'user_not_confirmed':
          return 'Email not confirmed yet. Check your inbox for the confirmation code, then confirm.'
        case 'code_mismatch':
          return 'That confirmation code is incorrect.'
        case 'code_expired':
          return 'That confirmation code has expired. Resend and try again.'
        case 'user_already_exists':
          return 'An account with that email already exists.'
        case 'not_authorized':
          return 'Invalid credentials, or the account is not allowed to sign in yet.'
        case 'rate_limited':
          return 'Too many attempts. Please wait and try again.'
        case 'invalid_password':
          return 'Password does not meet the pool requirements.'
        default:
          return code
      }
    }
    return `${err.status} ${err.message}`
  }
  if (err instanceof Error) return err.message
  return 'Request failed.'
}
