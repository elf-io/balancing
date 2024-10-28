#!/bin/bash

CURRENT_DIR_PATH=$(cd `dirname $0`; pwd)
PROJECT_ROOT_PATH="${CURRENT_DIR_PATH}/../.."
cd ${PROJECT_ROOT_PATH}

HEADER_TITLE="// Copyright $(date +%Y) Authors of elf-io"
SPDX_IDENTIFIER="// SPDX-License-Identifier: Apache-2.0"

GO_FILE_LIST=$(find . -type f -name "*.go" ! -path "./vendor/*" ! -path "./charts/*" ! -path "./output/*")

for FILE in $GO_FILE_LIST ; do
    # 检查文件的前两行是否包含完整的许可证头
    if ! head -n 2 "$FILE" | grep -q "${HEADER_TITLE}" || ! head -n 2 "$FILE" | grep -q "${SPDX_IDENTIFIER}"; then
        echo "inject license header to $FILE"
        echo -e "$HEADER_TITLE\n$SPDX_IDENTIFIER\n$(cat $FILE)" > $FILE
    fi
done
