#!/usr/bin/env bash
#
#   DbMan - © 2018-Present - SouthWinds Tech Ltd - www.southwinds.io
#   Licensed under the Apache License, Version 2.0 at http://www.apache.org/licenses/LICENSE-2.0
#   Contributors to this project, hereby assign copyright in this code to the project,
#   to be licensed under the same terms as the rest of the code.
#
VERSION=$1
if [ $# -eq 0 ]; then
    echo "An image version is required for DbMan. Provide it as a parameter."
    echo "Usage is: sh build.sh [APP VERSION] - e.g. sh build.sh v1.0.0"
    exit 1
fi

rm version

# creates a TAG for the newly built docker images
DATE=`date '+%d%m%y%H%M%S'`
HASH=`git rev-parse --short HEAD`
TAG="${VERSION}-${HASH}-${DATE}"

echo ${TAG} >> version

echo "TAG is: ${TAG}"

sleep 2