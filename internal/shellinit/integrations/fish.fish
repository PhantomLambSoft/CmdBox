function cbe --description 'Run a CmdBox command in the current shell'
    set -l tmpdir /tmp
    set -q TMPDIR; and set tmpdir (string trim --right --chars=/ -- $TMPDIR)
    set -l emit_file (mktemp $tmpdir/cmdbox_emit.XXXXXX)
    or return 1
    env CMDBOX_EMIT_FILE=$emit_file cb $argv --emit
    set -l rc $status
    if test $rc -eq 0; and test -s $emit_file
        source $emit_file
        set rc $status
    end
    command rm -f -- $emit_file
    return $rc
end
