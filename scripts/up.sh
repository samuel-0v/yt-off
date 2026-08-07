#!/usr/bin/env sh
set -eu

script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
project_dir="$(CDPATH= cd -- "$script_dir/.." && pwd)"

local_ip="${LOCAL_NETWORK_IP:-$("$script_dir/detect-local-ip.sh")}"
env_file="$project_dir/.env"
temp_file="$project_dir/.env.tmp"

if [ -f "$env_file" ]; then
	awk -v local_ip="$local_ip" '
		BEGIN { written = 0 }
		/^LOCAL_NETWORK_IP=/ {
			print "LOCAL_NETWORK_IP=" local_ip
			written = 1
			next
		}
		{ print }
		END {
			if (!written) {
				print "LOCAL_NETWORK_IP=" local_ip
			}
		}
	' "$env_file" > "$temp_file"
else
	printf 'LOCAL_NETWORK_IP=%s\n' "$local_ip" > "$temp_file"
fi

mv "$temp_file" "$env_file"
export LOCAL_NETWORK_IP="$local_ip"

printf 'yt-off local URL: http://%s:5173\n' "$local_ip"

cd "$project_dir"
docker compose up -d --build "$@"
