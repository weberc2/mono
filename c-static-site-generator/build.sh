#!/bin/bash

SCRIPTDIR="$(dirname $(realpath "$0"))"
BUILDDIR=$SCRIPTDIR/build
LIBDIR=$BUILDDIR/lib
TESTDIR=$BUILDDIR/tests
CC="clang -g -O0 -std=c2x -Wall"

libraries=""
function buildLibrary() {
    for dir in $SCRIPTDIR/src/*; do
        subdir=$(basename $dir)
        for dir in $SCRIPTDIR/src/$subdir/*; do
            if [[ -d $dir ]]; then
                lib=$(basename $dir)
                objdir=$LIBDIR/src/$lib
                mkdir -p $objdir

                (cd $objdir && $CC -I $SCRIPTDIR/include -c $(ls $dir/*.c | grep -v _test.c))
                ar -crs $LIBDIR/lib${lib}.a $objdir/*.o
                libraries="$libraries -l$lib"
            fi
        done
    done
}

tests=""
function buildPrivateTests() {
    for dir in $SCRIPTDIR/src/*; do 
        subdir=$(basename $dir)
        for dir in $SCRIPTDIR/src/$subdir/*; do
            if [[ -d $dir ]]; then
                lib=$(basename $dir)
                if [[ -n $(ls $dir/*_test.c 2> /dev/null) ]]; then
                    test=$BUILDDIR/src/$subdir/${lib}/tests
                    mkdir -p "$(dirname $test)"
                    $CC \
                        -I $SCRIPTDIR/include \
                        -L $LIBDIR $libraries \
                        -o $test \
                        $dir/*_test.c
                    tests="$tests && $test"
                fi
            fi
        done
    done
}

function buildTests() {
    mkdir -p $TESTDIR
    $CC \
        -I $SCRIPTDIR/include \
        -L $LIBDIR \
        $libraries \
        -o $TESTDIR/tests \
        $SCRIPTDIR/tests/*.c
}

function runTests() {
    bash -c "$TESTDIR/tests $tests"
}

set -eo pipefail
mkdir -p $BUILDDIR
mkdir -p $LIBDIR
buildLibrary
buildTests
buildPrivateTests
runTests