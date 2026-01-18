cbe() {
  local out status
  out="$(command cb "$@" --emit)"
  status=$?
  if [ $status -ne 0 ]; then
    return $status
  fi
  [ -z "$out" ] && return 0
  eval "$out"
}
