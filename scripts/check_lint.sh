#!/bin/bash

printerr() { printf "$@" 1>&2; }

PASS=$'\e[0;32m'
ERROR=$'\e[0;31m'
INFO=$'\e[0m'
WARN=$'\e[0;33m'

printf "Running command 'golangci-lint run --config ./.golangci.yml'"
`golangci-lint run --config ./.golangci.yml > lint.out`

if [ -s lint.out ]
then
    printerr "\n${ERROR}Lint check has failed.${INFO}\n"
    printerr "Please check the file ${WARN}lint.out${INFO}\n"
    exit 1
else
    printf "${PASS}\nLint check passed.${INFO}\n"
fi

exit 0