const API_BASE = import.meta.env.VITE_API_URL || '';

function getToken(): string | null { return localStorage.getItem('token'); }

function getAbsoluteBase(): string {
  if (API_BASE) return API_BASE;
  return `${window.location.protocol}//${window.location.host}`;
}

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
  const token = getToken();
  const headers: Record<string, string> = { 'Content-Type': 'application/json', ...(options.headers as Record<string, string> || {}) };
  if (token) headers['Authorization'] = `Bearer ${token}`;
  const base = API_BASE || '';
  const res = await fetch(`${base}${path}`, { ...options, headers });
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(err.error || res.statusText);
  }
  return res.json();
}

export const api = {
  login: (email: string, password: string) => request<{ user: any; token: string }>('/api/auth/login', { method: 'POST', body: JSON.stringify({ email, password }) }),
  register: (name: string, email: string, password: string, role: string) => request<{ user: any; token: string }>('/api/auth/register', { method: 'POST', body: JSON.stringify({ name, email, password, role }) }),
  me: () => request<any>('/api/auth/me'),
  getCases: (modality?: string, search?: string) => {
    const params = new URLSearchParams();
    if (modality) params.set('modality', modality);
    if (search) params.set('search', search);
    return request<any[]>(`/api/cases?${params}`);
  },
  getCase: (id: number) => request<any>(`/api/cases/${id}`),
  getModalities: () => request<string[]>('/api/cases/modalities'),
  createClassroom: (case_id: number, title: string) => request<any>('/api/classrooms', { method: 'POST', body: JSON.stringify({ case_id, title }) }),
  listClassrooms: () => request<any[]>('/api/classrooms'),
  joinClassroom: (code: string) => request<any>(`/api/classrooms/${code}`),
  endClassroom: (code: string) => request<any>(`/api/classrooms/${code}/end`, { method: 'PUT' }),
  mediaUrl: (folderName: string, phaseFolderName: string, fileName: string) => {
    const base = API_BASE || '';
    return `${base}/media/${folderName}/${phaseFolderName}/${fileName}`;
  },
  wsUrl: (code: string) => {
    const abs = getAbsoluteBase();
    const wsBase = abs.replace(/^http/, 'ws');
    return `${wsBase}/ws/classrooms/${code}?token=${getToken()}`;
  },
  getApiBase: () => API_BASE || '',
};
