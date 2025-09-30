export const API_BASE = (import.meta.env.VITE_API_BASE_URL as string | undefined) ?? '/api';

interface RequestOptions extends RequestInit {
  json?: unknown;
}

async function request<T>(path: string, options: RequestOptions = {}): Promise<T> {
  const url = `${API_BASE}${path}`;
  const headers = new Headers(options.headers ?? {});
  headers.set('Accept', 'application/json');

  let body: BodyInit | undefined = options.body ?? null;
  if (options.json !== undefined) {
    headers.set('Content-Type', 'application/json');
    body = JSON.stringify(options.json);
  } else if (body && !(body instanceof FormData) && typeof body !== 'string') {
    headers.set('Content-Type', 'application/json');
    body = JSON.stringify(body);
  }

  const response = await fetch(url, { ...options, headers, body: body ?? undefined });
  if (!response.ok) {
    const message = await extractErrorMessage(response);
    throw new Error(message);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  const contentType = response.headers.get('content-type') ?? '';
  if (contentType.includes('application/json')) {
    return (await response.json()) as T;
  }

  const text = await response.text();
  try {
    return JSON.parse(text) as T;
  } catch {
    return text as unknown as T;
  }
}

async function extractErrorMessage(response: Response): Promise<string> {
  try {
    const data = await response.clone().json();
    if (data && typeof data === 'object' && 'error' in data) {
      return String(data.error);
    }
  } catch {
    // ignore json parse errors
  }
  const text = await response.text();
  if (text.trim().length > 0) {
    return text;
  }
  return `HTTP ${response.status}`;
}

export function getJson<T>(path: string, options?: RequestOptions): Promise<T> {
  return request<T>(path, { ...options, method: options?.method ?? 'GET' });
}

export function postJson<T>(path: string, body?: unknown, options?: RequestOptions): Promise<T> {
  return request<T>(path, { ...options, method: 'POST', json: body });
}

export function putJson<T>(path: string, body?: unknown, options?: RequestOptions): Promise<T> {
  return request<T>(path, { ...options, method: 'PUT', json: body });
}

export function deleteJson(path: string, options?: RequestOptions): Promise<void> {
  return request<void>(path, { ...options, method: 'DELETE' });
}
