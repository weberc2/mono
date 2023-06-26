#!/bin/bash

SCRIPTDIR="$(dirname $(realpath "$0"))"
BUILDDIR=$SCRIPTDIR/build
LIBDIR=$BUILDDIR/lib
CC="clang -g -O0 -std=c11 -Wall"

libraries=""
function buildLibrary() {
    subdir="std"
    for dir in $SCRIPTDIR/src/$subdir/*; do
        if [[ -d $dir ]]; then
            lib=$(basename $dir)
            objdir=$LIBDIR/src/$lib
            mkdir -p $objdir

            (cd $objdir && $CC -I $SCRIPTDIR/include -c $dir/*.c)
            ar -crs $LIBDIR/lib${lib}.a $objdir/*.o
            libraries="$libraries -l$lib"
        fi
    done

    subdir="core"
    for dir in $SCRIPTDIR/src/$subdir/*; do
        if [[ -d $dir ]]; then
            lib=$(basename $dir)
            objdir=$LIBDIR/src/$subdir/$lib
            mkdir -p $objdir

            (cd $objdir && $CC -I $SCRIPTDIR/include -c $dir/*.c)
            ar -crs $LIBDIR/lib${lib}.a $objdir/*.o
            libraries="$libraries -l$lib"
        fi
    done
}

# function buildBinary() {
#     bindir=$BUILDDIR/bin
#     mkdir -p $bindir
#     $CC -I $SCRIPTDIR/include -L $LIBDIR $libraries $SCRIPTDIR/src/main.c -o $bindir/ctest
# }

function buildTests() {
    testsdir=$BUILDDIR/tests
    mkdir -p $testsdir
    $CC \
        -I $SCRIPTDIR/include \
        -L $LIBDIR \
        $libraries \
        -o $testsdir/tests \
        $SCRIPTDIR/tests/*.c
}

set -eo pipefail
mkdir -p $BUILDDIR
mkdir -p $LIBDIR
buildLibrary
buildTests