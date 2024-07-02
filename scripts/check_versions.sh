#!/bin/bash

set -e

printerr() { echo "$@" 1>&2; }

PASS=$'\e[0;32m'
ERROR=$'\e[0;31m'
INFO=$'\e[0m'
WARN=$'\e[0;33m'



if [ -z "$GOPATH" ]
then
    printerr "\${ERROR}$GOPATH is empty, please set in for your environment.\n"
    exit 1
else

    # CHECK GO VERSION- THROW ONLY WARNING IF IT DOES NOT MATCH WITH THE DESIRED VERSION 
    printf "Checking go version at ${GOPATH}\n"

    GOVERSION=`go version`
    printf $GOVERSION
    if [[ $GOVERSION == *"1.22."* ]]
    then
        printf "${PASS}Go version is correctly set in your environment${INFO}\n"
    else
        printerr "${WARN}$GOVERSION is not the ideal Go version. Please set it to${GREEN}1.22.x${INFO}\n"
    fi

    # CHECK GOLANG-CI LINT VERSION - THROW ONLY WARNING IF IT DOES NOT MATCH WITH THE DESIRED VERSION
    printf "Checking golangci-lint version at ${GOPATH}/bin/golangci-lint\n"
    LINTVERSION=`golangci-lint --version`
    printf $LINTVERSION
    if [[ $LINTVERSION == *"1.55.2"* ]]
    then
        printf "${PASS}golangci-lint version is correctly set in your environment${INFO}\n"
    else
        printerr "${WARN}$LINTVERSION is not the ideal golangci-lint version. Please set it to${GREEN}1.55.2${INFO}\n"
    fi

    exit 0
fi
