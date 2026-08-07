FROM python:3.12-alpine

RUN apk add --no-cache ffmpeg nodejs \
  && pip install --no-cache-dir yt-dlp

WORKDIR /downloads

CMD ["tail", "-f", "/dev/null"]
