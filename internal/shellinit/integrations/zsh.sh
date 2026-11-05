cbe() {
  local _cb_file _cb_rc
  _cb_file="$(mktemp "${TMPDIR:-/tmp}/cmdbox_emit.XXXXXX")" || return 1
  CMDBOX_EMIT_FILE="$_cb_file" command cb "$@" --emit
  _cb_rc=$?
  if [ "$_cb_rc" -eq 0 ] && [ -s "$_cb_file" ]; then
    . "$_cb_file"
    _cb_rc=$?
  fi
  command rm -f -- "$_cb_file"
  return "$_cb_rc"
}
