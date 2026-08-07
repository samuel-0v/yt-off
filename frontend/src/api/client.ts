import type {
  CreateDownloadResponse,
  DownloadGroup,
  DownloadGroupItem,
  DownloadTask,
  FileInfo,
  User,
  VideoInfo,
} from '../types/download'
import type {
  AppSettings,
  CookiesInfo,
  DockerStatus,
  HealthStatus,
  NetworkInfo,
  StorageInfo,
  SystemInfo,
  VersionInfo,
  YTDLPVersionInfo,
} from '../types/system'

const API_BASE_URL = normalizeBaseUrl(import.meta.env.VITE_API_BASE_URL ?? '/api')
const ROOT_BASE_URL = getRootBaseUrl(API_BASE_URL)

type ApiErrorBody = {
  error?: string
}

export class ApiClientError extends Error {
  status: number

  constructor(message: string, status: number) {
    super(message)
    this.name = 'ApiClientError'
    this.status = status
  }
}

export async function getHealth(): Promise<HealthStatus> {
  return requestFromRoot<HealthStatus>('/health')
}

export async function getVersion(): Promise<VersionInfo> {
  return request<VersionInfo>('/version')
}

export async function getStorage(): Promise<StorageInfo> {
  return request<StorageInfo>('/storage')
}

export async function getDockerStatus(): Promise<DockerStatus> {
  return request<DockerStatus>('/docker/status')
}

export async function getSettings(): Promise<AppSettings> {
  return request<AppSettings>('/settings')
}

export async function updateSettings(settings: AppSettings): Promise<AppSettings> {
  return request<AppSettings>('/settings', {
    method: 'PUT',
    body: JSON.stringify(settings),
  })
}

export async function getSystemInfo(): Promise<SystemInfo> {
  return request<SystemInfo>('/system')
}

export async function getYTDLPVersion(): Promise<YTDLPVersionInfo> {
  return request<YTDLPVersionInfo>('/ytdlp/version')
}

export async function getNetworkInfo(): Promise<NetworkInfo> {
  return request<NetworkInfo>('/network')
}

export async function getCookiesInfo(): Promise<CookiesInfo> {
  return request<CookiesInfo>('/cookies')
}

export async function saveCookies(content: string): Promise<CookiesInfo> {
  return request<CookiesInfo>('/cookies', {
    method: 'PUT',
    body: JSON.stringify({ content }),
  })
}

export async function uploadCookies(file: File): Promise<CookiesInfo> {
  const formData = new FormData()
  formData.append('file', file)

  return request<CookiesInfo>('/cookies', {
    method: 'POST',
    body: formData,
    headers: {},
  })
}

export async function deleteCookies(): Promise<void> {
  return request<void>('/cookies', {
    method: 'DELETE',
  })
}

export async function getFormats(url: string): Promise<VideoInfo> {
  return request<VideoInfo>('/formats', {
    method: 'POST',
    body: JSON.stringify({ url }),
  })
}

export async function listUsers(): Promise<User[]> {
  return request<User[]>('/users')
}

export async function createUser(username: string): Promise<User> {
  return request<User>('/users', {
    method: 'POST',
    body: JSON.stringify({ username }),
  })
}

export async function createDownload(
  url: string,
  formatId: string,
  extension?: string,
  userId?: string,
): Promise<CreateDownloadResponse> {
  return request<CreateDownloadResponse>('/downloads', {
    method: 'POST',
    body: JSON.stringify({ url, format_id: formatId, extension, user_id: userId }),
  })
}

export async function getDownload(id: string): Promise<DownloadTask> {
  return request<DownloadTask>(`/downloads/${encodeURIComponent(id)}`)
}

export async function cancelDownload(id: string): Promise<DownloadTask> {
  return request<DownloadTask>(`/downloads/${encodeURIComponent(id)}`, {
    method: 'DELETE',
  })
}

export async function copyDownload(id: string, userId?: string): Promise<DownloadTask> {
  return request<DownloadTask>(`/downloads/${encodeURIComponent(id)}/copy`, {
    method: 'POST',
    body: JSON.stringify({ user_id: userId }),
  })
}

export async function listDownloads(scope: 'mine' | 'all' = 'all', userId?: string): Promise<DownloadTask[]> {
  return request<DownloadTask[]>(withQuery('/downloads', {
    scope: scope === 'mine' ? 'mine' : '',
    user_id: scope === 'mine' ? userId : '',
  }))
}

export async function listGroups(scope: 'mine' | 'all' = 'all', userId?: string): Promise<DownloadGroup[]> {
  return request<DownloadGroup[]>(withQuery('/groups', {
    scope: scope === 'mine' ? 'mine' : '',
    user_id: scope === 'mine' ? userId : '',
  }))
}

export async function getGroup(id: string): Promise<DownloadGroup> {
  return request<DownloadGroup>(`/groups/${encodeURIComponent(id)}`)
}

export async function createGroup(userId: string, name: string, description: string): Promise<DownloadGroup> {
  return request<DownloadGroup>('/groups', {
    method: 'POST',
    body: JSON.stringify({ user_id: userId, name, description }),
  })
}

export async function updateGroup(
  id: string,
  userId: string,
  name: string,
  description: string,
): Promise<DownloadGroup> {
  return request<DownloadGroup>(`/groups/${encodeURIComponent(id)}`, {
    method: 'PUT',
    body: JSON.stringify({ user_id: userId, name, description }),
  })
}

export async function deleteGroup(id: string, userId: string): Promise<void> {
  return request<void>(withQuery(`/groups/${encodeURIComponent(id)}`, { user_id: userId }), {
    method: 'DELETE',
  })
}

export async function addGroupItem(groupId: string, userId: string, downloadId: string): Promise<DownloadGroupItem> {
  return request<DownloadGroupItem>(`/groups/${encodeURIComponent(groupId)}/items`, {
    method: 'POST',
    body: JSON.stringify({ user_id: userId, download_id: downloadId }),
  })
}

export async function removeGroupItem(groupId: string, itemId: string, userId: string): Promise<void> {
  return request<void>(withQuery(`/groups/${encodeURIComponent(groupId)}/items/${encodeURIComponent(itemId)}`, {
    user_id: userId,
  }), {
    method: 'DELETE',
  })
}

export async function listFiles(): Promise<FileInfo[]> {
  return request<FileInfo[]>('/files')
}

export async function deleteFile(name: string): Promise<void> {
  return request<void>(`/files/${encodeURIComponent(name)}`, {
    method: 'DELETE',
  })
}

export function getFileDownloadURL(name: string): string {
  return `${API_BASE_URL}/files/${encodeURIComponent(name)}`
}

export function getFilePreviewURL(name: string): string {
  return `${getFileDownloadURL(name)}?inline=1`
}

async function request<T>(path: string, init: RequestInit = {}): Promise<T> {
  return requestWithBase<T>(API_BASE_URL, path, init)
}

async function requestFromRoot<T>(path: string, init: RequestInit = {}): Promise<T> {
  return requestWithBase<T>(ROOT_BASE_URL, path, init)
}

async function requestWithBase<T>(baseUrl: string, path: string, init: RequestInit = {}): Promise<T> {
  const headers = init.body instanceof FormData
    ? init.headers
    : {
        'Content-Type': 'application/json',
        ...init.headers,
      }

  const response = await fetch(`${baseUrl}${path}`, {
    ...init,
    headers,
  })

  if (!response.ok) {
    const body = await readErrorBody(response)
    throw new ApiClientError(body.error ?? 'request failed', response.status)
  }

  if (response.status === 204) {
    return undefined as T
  }

  const text = await response.text()
  if (!text) {
    return undefined as T
  }

  return JSON.parse(text) as T
}

async function readErrorBody(response: Response): Promise<ApiErrorBody> {
  try {
    return (await response.json()) as ApiErrorBody
  } catch {
    return {}
  }
}

function normalizeBaseUrl(value: string): string {
  return value.endsWith('/') ? value.slice(0, -1) : value
}

function getRootBaseUrl(value: string): string {
  if (value.endsWith('/api')) {
    return value.slice(0, -4)
  }

  return value
}

function withQuery(path: string, params: Record<string, string | undefined>): string {
  const search = new URLSearchParams()
  Object.entries(params).forEach(([key, value]) => {
    if (value) {
      search.set(key, value)
    }
  })

  const query = search.toString()
  return query ? `${path}?${query}` : path
}
