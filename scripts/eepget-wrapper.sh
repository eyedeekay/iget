#! /usr/bin/env sh

echo "WARNING: YOU SHOULD UPDATE YOUR SCRIPTS. This wrapper is provided for your
convenience. If you are using it, it means that you have replaced the real
eepget with this wrapper. It is automatically substituting settings, and this
is the only warning it can give you." 1>&2

argslist="$@"

sam="127.0.0.1:7656"

while getopts ":p:l:" o; do
    case "${o}" in
        p)
            p="$OPTARG"
            ;;
        proxy_host)
            p="$OPTARG"
            ;;
        l)
            l="$OPTARG"
            ;;
        lineLen)
            l="$OPTARG"
            ;;
        *)
            ;;
    esac
done
shift $((OPTIND-1))



if [ ! -z $p ]; then
    if [ "$p" = "127.0.0.1:4444" ]; then
        args=$(echo "$argslist" | sed "s|$p|$sam|g" 2>&1 | tr ' ' ' ')
    elif [ "$p" = "localhost:4444" ]; then
        args=$(echo "$argslist" | sed "s|$p|$sam|g" 2>&1 | tr ' ' ' ')
    elif [ "$p" = "http://127.0.0.1:4444" ]; then
        args=$(echo "$argslist" | sed "s|$p|$sam|g" 2>&1 | tr ' ' ' ')
    elif [ "$p" = "http://localhost:4444" ]; then
        args=$(echo "$argslist" | sed "s|$p|$sam|g" 2>&1 | tr ' ' ' ')
    fi
else
    args="$argslist"
fi

if [ ! -z "${l}" ]; then
    argsf=$(echo $args | sed "s|-l $l||g" 2>&1 | sed "s|-lineLen $1||g" 2>&1 | tr ' ' ' ')
    $(which iget) $argsf | fold -w "$l" -s -
else
    $(which iget) $args
fi
