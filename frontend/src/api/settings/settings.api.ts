import { requestJson } from '../http'
import type { UserSettings } from './settings.types'

export const settingsApi = {
  async get(userType: string): Promise<UserSettings> {
    return requestJson<UserSettings>({
      method: 'GET',
      path: '/api/settings',
      headers: { 'X-User-Type': userType },
    })
  },

  async update(settings: UserSettings, userType: string): Promise<UserSettings> {
    return requestJson<UserSettings>({
      method: 'PUT',
      path: '/api/settings',
      headers: { 'X-User-Type': userType },
      body: settings,
    })
  },
}
