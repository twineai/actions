#!/bin/bash -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
SCRIPT_NAME=$(basename "${BASH_SOURCE[0]}")

echoerr() {
  >&2 echo "$@"
}

usage() {
  cat << EOF
Usage:
  ${SCRIPT_NAME} ACTION_DIR[...]

Arguments
  ACTION_DIR The path to a directory containing an action. May be
    specified multiple times.
EOF
}

if [[ $# -lt 1 ]]; then
  echoerr "no actions provided"
  usage
  exit 1
fi

TARCMD=""
case "$(uname -s)" in
    Linux*)
      TARCMD=$(which tar)
      ;;
    Darwin*)
      TARCMD=$(which gtar)
      if [[ -z "${TARCMD}" ]]; then
        echoerr "This command requires GNU Tar. Please install it via `brew install gnu-tar`"
        exit 1
      fi
      ;;
esac

BASE_DIR=$(pwd)

for RAW_ACTION in "$@"; do
  ACTION="${RAW_ACTION%/}"
  echo "Packaging action: ${ACTION}"

  cd "${BASE_DIR}"

  if [[ ! -d "${ACTION}" ]]; then
    echoerr "action ${ACTION} is not a directory"
    usage
    exit 1
  fi

  cd "${ACTION}"
  find . \( -name node_modules -o -name .DS_Store \) -prune -o -type f -exec \
    echo ${TARCMD} -C "${ACTION}" \
        --sort=name \
        --mtime="1970-01-01" \
        --owner=0 --group=0 --numeric-owner \
        -cf - {} \+ | gzip -n > "${BASE_DIR}/${ACTION}.tgz"
done