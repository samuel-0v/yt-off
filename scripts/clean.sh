#!/usr/bin/env sh
set -eu

script_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
project_dir="$(CDPATH= cd -- "$script_dir/.." && pwd)"
downloads_dir="$project_dir/downloads"
assume_yes=0

for arg in "$@"; do
	case "$arg" in
		-y|--yes)
			assume_yes=1
			;;
		*)
			printf 'Unknown option: %s\n' "$arg" >&2
			printf 'Usage: %s [--yes]\n' "$0" >&2
			exit 2
			;;
	esac
done

if [ "$assume_yes" -ne 1 ]; then
	printf 'This will stop yt-off, remove the SQLite database volume, and delete all files in downloads/.\n'
	printf 'Cookies in cookies/ will be preserved.\n'
	printf 'Continue? [y/N] '
	read -r answer
	case "$answer" in
		y|Y|yes|YES)
			;;
		*)
			printf 'Cleanup cancelled.\n'
			exit 0
			;;
	esac
fi

cd "$project_dir"

printf 'Stopping containers and removing database volume...\n'
docker compose down -v

printf 'Cleaning downloads directory...\n'
mkdir -p "$downloads_dir"
find "$downloads_dir" -mindepth 1 -maxdepth 1 ! -name .gitkeep -exec rm -rf {} +
touch "$downloads_dir/.gitkeep"

printf 'Cleanup complete. Run ./scripts/up.sh or docker compose up -d --build to start yt-off again.\n'
