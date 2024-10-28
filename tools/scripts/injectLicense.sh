#!/bin/bash

CURRENT_DIR_PATH=$(cd `dirname $0`; pwd)
PROJECT_ROOT_PATH="${CURRENT_DIR_PATH}/../.."
cd ${PROJECT_ROOT_PATH}

HEADER_TITLE="// Copyright $(date +%Y) Authors of elf-io\n// SPDX-License-Identifier: Apache-2.0"

GO_FILE_LIST=$( grep -v "${HEADER_TITLE}" --include \*.go  . -RHn --colour -l  --exclude-dir={vendor,charts,output} )

for FILE in $GO_FILE_LIST ; do
   echo "inject license header to $FILE"
   sed -i '1i\
   '$HEADER_TITLE'
   ' $FILE
done

