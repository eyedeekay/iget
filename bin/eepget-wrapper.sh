#! /usr/bin/env sh
while [[ $# > 1 ]]
do
key="$1"

case $key in
    -p|--proxy)
    DISREGARD_PROXY="$2"
    shift
    ;;
    *)
    NEW_ARGS="$2"
    shift
    ;;
esac
shift # past argument or value
done

if "x$DISREGARD_PROXY" != "x"; then
    echo "Disregarding the http proxy $DISREGARD_PROXY, this uses SAM"
fi

NEW_ARGS=""

iget $@ #| fold -w 80 -s -
