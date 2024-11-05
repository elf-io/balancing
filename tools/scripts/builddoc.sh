#!/bin/bash

pip install mkdocs mkdocs-material mkdocs-i18n

CURRENT_DIR_PATH=$(cd `dirname $0`; pwd)
PROJECT_ROOT_PATH="${CURRENT_DIR_PATH}/../.."

cd ${PROJECT_ROOT_PATH}

echo "in $(pwd)"

echo "-----------"
cp ./docs/mkdocs.yml ./
mkdocs build
rm -f mkdocs.yml || true
