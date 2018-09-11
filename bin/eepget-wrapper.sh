#! /usr/bin/env sh

echo "WARNING: YOU SHOULD UPDATE YOUR SCRIPTS. This wrapper is provided for your
convenience. If you are using it, it means that you have replaced the real
eepget with this wrapper. It is automatically substituting settings, and this
is the only warning it can give you." 1>&2

argslist=$@

while getopts ":p:l:" o; do
    case "${o}" in
        p)
            p=${OPTARG}
            ;;
        proxy_host)
            p=${OPTARG}
            ;;
        l)
            l=${OPTARG}
            ;;
        lineLen)
            l=${OPTARG}
            ;;
        *)
            ;;
    esac
done
shift $((OPTIND-1))

if [ ! -z "${p}" ]; then
    if [ "$p" = "127.0.0.1:4444" ]; then
        p="127.0.0.1:7656"
        args=$(echo $argslist | sed "s|127.0.0.1:4444|$p|g" )
    elif [ "$p" = "localhost:4444" ]; then
        p="localhost:7656"
        args=$(echo $argslist | sed "s|localhost:4444|$p|g" )
    elif [ "$p" = "http://127.0.0.1:4444" ]; then
        p="http://127.0.0.1:7656"
        args=$(echo $argslist | sed "s|http://127.0.0.1:4444|$p|g" )
    elif [ "$p" = "http://localhost:4444" ]; then
        p="http://localhost:7656"
        args=$(echo $argslist | sed "s|http://localhost:4444|$p|g" )
    fi
fi

if [ ! -z "${l}" ]; then
    argsf=$(echo $args | sed "s|-l $l||g" | sed "s|-lineLen $1||g")
    $(which iget) $argsf | fold -w "$l" -s -
else
    $(which iget) $args
fi
