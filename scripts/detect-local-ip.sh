#!/usr/bin/env sh
set -eu

detect_with_ip_route() {
	if ! command -v ip >/dev/null 2>&1; then
		return 1
	fi

	ip -4 route get 1.1.1.1 2>/dev/null | awk '{
		for (i = 1; i <= NF; i++) {
			if ($i == "src") {
				print $(i + 1)
				exit
			}
		}
	}'
}

detect_with_hostname() {
	if ! command -v hostname >/dev/null 2>&1; then
		return 1
	fi

	hostname -I 2>/dev/null | awk '{print $1}'
}

local_ip="$(detect_with_ip_route || true)"
if [ -z "$local_ip" ]; then
	local_ip="$(detect_with_hostname || true)"
fi

if [ -z "$local_ip" ]; then
	echo "Could not detect local IP. Set LOCAL_NETWORK_IP manually." >&2
	exit 1
fi

printf '%s\n' "$local_ip"
