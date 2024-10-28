#!/bin/bash

CURRENT_DIR_PATH=$(cd `dirname $0`; pwd)
PROJECT_ROOT_PATH="${CURRENT_DIR_PATH}/../.."
cd ${PROJECT_ROOT_PATH}

HEADER_TITLE="// Copyright $(date +%Y) Authors of elf-io\n// SPDX-License-Identifier: Apache-2.0"

GO_FILE_LIST=$( grep -v "${HEADER_TITLE}" --include \*.go  . -RHn --colour -l  --exclude-dir={vendor,charts,output} )

echo -e "$HEADER_TITLE" > /tmp/.header

for FILE in $GO_FILE_LIST ; do
   echo "inject license header to $FILE"
   while read LINE ; do
       # 使用空的备份后缀以确保兼容性
       sed -i '' '1i  '$LINE'' $FILE
   done < /tmp/.header
done
