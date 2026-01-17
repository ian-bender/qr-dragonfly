import { requestJson } from '../http';

export interface CreateCheckoutSessionRequest {
  plan: 'basic' | 'enterprise';
}

export interface CreateCheckoutSessionResponse {
  sessionId: string;
  url: string;
}

export interface CreatePortalSessionResponse {
  url: string;
}

export async function createCheckoutSession(
  plan: 'basic' | 'enterprise'
): Promise<CreateCheckoutSessionResponse> {
  return requestJson<CreateCheckoutSessionResponse>({
    path: '/api/stripe/checkout-session',
    method: 'POST',
    body: { plan },
  });
}

export async function createPortalSession(): Promise<CreatePortalSessionResponse> {
  return requestJson<CreatePortalSessionResponse>({
    path: '/api/stripe/portal-session',
    method: 'POST',
  });
}
