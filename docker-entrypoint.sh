#!/bin/sh
set -e

# Require git identity environment variables
if [ -z "$GIT_USER_NAME" ]; then
    echo "ERROR: GIT_USER_NAME environment variable is required." >&2
    echo "  Usage: docker run -e GIT_USER_NAME=\"Your Name\" -e GIT_USER_EMAIL=\"you@example.com\" ..." >&2
    exit 1
fi

if [ -z "$GIT_USER_EMAIL" ]; then
    echo "ERROR: GIT_USER_EMAIL environment variable is required." >&2
    echo "  Usage: docker run -e GIT_USER_NAME=\"Your Name\" -e GIT_USER_EMAIL=\"you@example.com\" ..." >&2
    exit 1
fi

git config --global user.name "$GIT_USER_NAME"
git config --global user.email "$GIT_USER_EMAIL"

exec /usr/local/bin/release-it-go "$@"
