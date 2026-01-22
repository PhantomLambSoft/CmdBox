function cbe
    set -l out (cb cmdbox $argv --emit)
    set -l status $status
    if test $status -ne 0
        return $status
    end
    if test -z "$out"
        return 0
    end
    eval $out
end
