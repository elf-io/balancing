#!/bin/bash


set -o errexit
set -o nounset
set -o pipefail

CURRENT_DIR_PATH=$(cd `dirname $0`; pwd)
PROJECT_ROOT_PATH="${CURRENT_DIR_PATH}/../.."
cd ${PROJECT_ROOT_PATH}

HEADER_TITLE="Copyright $(date +%Y) Authors of elf-io"
SPDX_IDENTIFIER="SPDX-License-Identifier: Apache-2.0"

GO_FILE_LIST=$(find . -type f -name "*.go" ! -path "./vendor/*" ! -path "./charts/*" ! -path "./output/*")
for FILE in $GO_FILE_LIST ; do
    # 检查文件的前两行是否包含完整的许可证头
    if ! grep "SPDX-License-Identifier" $FILE &>/dev/null ; then
        echo "inject license header to $FILE"
        echo -e "// $HEADER_TITLE\n// $SPDX_IDENTIFIER\n$(cat $FILE)" > $FILE
    fi
done

SH_FILE_LIST=$(find . -type f -name "*.sh" ! -path "./vendor/*" ! -path "./charts/*" ! -path "./output/*")
for FILE in ${SH_FILE_LIST} ; do
    if ! grep "SPDX-License-Identifier" $FILE &>/dev/null ; then
        echo "inject license header to $FILE"
        sed -i '1 a \# '"${SPDX_IDENTIFIER}"'' $FILE
        sed -i '1 a \# '"${HEADER_TITLE}"'' $FILE
    fi
done
