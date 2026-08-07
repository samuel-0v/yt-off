export type HealthStatus = {
  status: string
}

export type VersionInfo = {
  name: string
  version: string
  environment: string
}

export type StorageInfo = {
  total: number
  used: number
  free: number
  usage_percent: number
}

export type DockerStatus = {
  docker: string
  ytdlp_container: string
}

export type AppSettings = {
  download_directory: string
  max_concurrent_downloads: number
  language: string
  theme: 'system' | 'light' | 'dark'
  app_name: string
  backend_port: string
  automatic_updates: boolean
  show_hidden_files: boolean
}

export type SystemInfo = {
  os: string
  architecture: string
  docker: string
  sqlite: string
  backend_version: string
  yt_dlp_version: string
  ffmpeg_version: string
}

export type YTDLPVersionInfo = {
  current: string
}

export type NetworkInfo = {
  hostname: string
  local_ip: string
  backend_port: string
  frontend_port: string
  url: string
}

export type CookiesInfo = {
  exists: boolean
  valid: boolean
  file_name: string
  size?: number
  updated_at?: string
  message?: string
}
