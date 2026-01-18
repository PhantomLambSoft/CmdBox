function cb
    set -l out (command cmdbox $argv --emit)
    set -l status $status
    if test $status -ne 0
        return $status
    end
    if test -z "$out"
        return 0
    end
    eval $out
end
