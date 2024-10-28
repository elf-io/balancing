#!/bin/bash

HEADER_TITLE="// Copyright $(date +%Y) Authors of elf-io\n// SPDX-License-Identifier: Apache-2.0"

GO_FILE_LIST=$( grep -v "${HEADER_TITLE}" --include \*.go  . -RHn --colour -l  --exclude-dir={vendor,charts,output} )

for FILE in $GO_FILE_LIST ; do
   echo "inject license header to $FILE"
   sed -i '1i\
   '$HEADER_TITLE'
   ' $FILE
done

