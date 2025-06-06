#!/bin/bash

# Copyright 2024 Authors of kdoctor-io
# SPDX-License-Identifier: Apache-2.0

pip install mkdocs mkdocs-material mkdocs-i18n mkdocs-material-extensions

CURRENT_DIR_PATH=$(cd `dirname $0`; pwd)
PROJECT_ROOT_PATH="${CURRENT_DIR_PATH}/../.."

cd ${PROJECT_ROOT_PATH}

echo "in $(pwd)"

echo "-----------"
cp ./docs/mkdocs.yml ./
mkdocs build
rm -f mkdocs.yml || true
