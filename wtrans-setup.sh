#!/bin/sh
set -eu

BIN="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)/wtrans"
if [ ! -x "$BIN" ]; then
  echo "wtrans binary was not found next to this script."
  exit 1
fi

while :; do
  printf '\n'
  printf '%s\n' "wtrans quick setup"
  printf '%s\n' "1. List visible windows"
  printf '%s\n' "2. Make an app transparent"
  printf '%s\n' "3. Turn transparent mode off for an app"
  printf '%s\n' "4. Show status"
  printf '%s\n' "5. Stop background keeper"
  printf '%s\n' "6. Reset all saved rules"
  printf '%s\n' "7. Exit"
  printf '%s' "Choose 1-7: "
  IFS= read -r CHOICE || exit 0

  case "$CHOICE" in
    1)
      printf '%s' "Process name (blank for all): "
      IFS= read -r PROCESS || PROCESS=""
      "$BIN" list --process "$PROCESS"
      ;;
    2)
      printf '%s' "Process name (example: WindowsTerminal.exe): "
      IFS= read -r PROCESS || PROCESS=""
      printf '%s' "Opacity 20-100 (example: 85): "
      IFS= read -r OPACITY || OPACITY=""
      printf '%s' "Keep future windows too? (y/n): "
      IFS= read -r PERSIST || PERSIST=""
      if [ "${PERSIST:-}" = "y" ] || [ "${PERSIST:-}" = "Y" ]; then
        "$BIN" set --process "$PROCESS" --opacity "$OPACITY" --persist
      else
        "$BIN" set --process "$PROCESS" --opacity "$OPACITY"
      fi
      ;;
    3)
      printf '%s' "Process name: "
      IFS= read -r PROCESS || PROCESS=""
      "$BIN" restore --process "$PROCESS"
      ;;
    4)
      "$BIN" status
      ;;
    5)
      "$BIN" stop
      ;;
    6)
      printf '%s' "This will clear all saved rules. Type RESET to continue: "
      IFS= read -r CONFIRM || CONFIRM=""
      if [ "${CONFIRM:-}" = "RESET" ]; then
        "$BIN" reset
      fi
      ;;
    7)
      exit 0
      ;;
  esac
done
