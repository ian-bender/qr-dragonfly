import { requestJson } from '../http'
import type { User } from '../users/users.types'

export interface AdminUser extends User {
  createdAt?: string
  enabled?: boolean
  status?: string
}

export interface UpdateUserRequest {
  email?: string
  userType?: string
}

export const adminApi = {
  async listUsers(adminKey: string): Promise<AdminUser[]> {
    return requestJson<AdminUser[]>({
      method: 'GET',
      path: '/api/users',
      headers: { 'X-Admin-Key': adminKey },
    })
  },

  async getUser(userId: string, adminKey: string): Promise<AdminUser> {
    return requestJson<AdminUser>({
      method: 'GET',
      path: `/api/users/${encodeURIComponent(userId)}`,
      headers: { 'X-Admin-Key': adminKey },
    })
  },

  async updateUser(
    userId: string,
    updates: UpdateUserRequest,
    adminKey: string
  ): Promise<AdminUser> {
    return requestJson<AdminUser>({
      method: 'PATCH',
      path: `/api/users/${encodeURIComponent(userId)}`,
      headers: { 'X-Admin-Key': adminKey },
      body: updates as any,
    })
  },

  async deleteUser(userId: string, adminKey: string): Promise<void> {
    return requestJson<void>({
      method: 'DELETE',
      path: `/api/users/${encodeURIComponent(userId)}`,
      headers: { 'X-Admin-Key': adminKey },
    })
  },
}
