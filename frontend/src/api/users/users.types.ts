export type User = {
  id: string
  email: string
  name?: string
  userType?: 'free' | 'basic' | 'enterprise' | 'admin'
  createdAtIso?: string
}

export type CreateUserInput = {
  email: string
  password: string
  name?: string
}

export type UpdateUserInput = {
  email?: string
  password?: string
  name?: string
}

export type LoginInput = {
  email: string
  password: string
}

export type ConfirmSignUpInput = {
  email: string
  code: string
}

export type ResendConfirmationInput = {
  email: string
}

export type ForgotPasswordInput = {
  email: string
}

export type ConfirmForgotPasswordInput = {
  email: string
  code: string
  newPassword: string
}

export type ChangePasswordInput = {
  oldPassword: string
  newPassword: string
}

export type StatusResponse = {
  status: string
  delivery?: {
    destination?: string
    medium?: string
    attribute?: string
  }
}

export type AuthSession = {
  user: User
  token?: string
}
