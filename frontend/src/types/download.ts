export type DownloadStatus = 'queued' | 'running' | 'completed' | 'failed' | 'cancelled'

export type DownloadOption = {
  label: string
  format_id: string
  quality: string
  extension: string
  type: string
  resolution?: string
  has_video: boolean
  has_audio: boolean
  estimated_size?: number
  audio_codec?: string
  video_codec?: string
  bitrate?: number
}

export type VideoInfo = {
  title: string
  duration: number
  thumbnail: string
  channel?: string
  uploader?: string
  options: DownloadOption[]
}

export type CreateDownloadResponse = {
  id: string
  status: DownloadStatus
}

export type User = {
  id: string
  username: string
  created_at: string
  updated_at: string
}

export type DownloadTask = {
  id: string
  user_id?: string
  owner_username?: string
  url?: string
  format_id?: string
  status: DownloadStatus
  progress: number
  speed?: string
  eta?: string
  filename?: string
  file_size?: number
  extension?: string
  container_id?: string
  error?: string
  created_at?: string
  updated_at?: string
}

export type FileInfo = {
  name: string
  size: number
  extension: string
  modified_at: string
}

export type DownloadGroupItem = {
  id: string
  group_id: string
  download_id: string
  position: number
  download?: DownloadTask
  created_at: string
}

export type DownloadGroup = {
  id: string
  user_id: string
  owner_username?: string
  name: string
  description?: string
  item_count: number
  items?: DownloadGroupItem[]
  created_at: string
  updated_at: string
}
