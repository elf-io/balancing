# Copyright 2024 Authors of elf-io
# SPDX-License-Identifier: Apache-2.0

# ====modify====
ARG BASE_IMAGE=docker.io/library/ubuntu:24.10
ARG GOLANG_IMAGE=docker.io/library/golang:1.23.1

#========= build gops ===============
FROM --platform=${TARGETPLATFORM} ${GOLANG_IMAGE} as gops-cni-builder
COPY /images/build-gops.sh /build-gops.sh

ARG USE_PROXY_SOURCE
COPY /images/sources-debian12.list /sources.list
RUN [ "$USE_PROXY_SOURCE" != "true" ] && exit 0 ; \
      rm -f /etc/apt/sources.list ; \
      rm -rf /etc/apt/sources.list.d ; \
      cp /sources.list /etc/apt/sources.list

RUN apt-get update && apt-get install -y binutils-aarch64-linux-gnu binutils-x86-64-linux-gnu

# build-gops.sh will build both archs at the same time
WORKDIR /
RUN chmod +x /build-gops.sh && /build-gops.sh

#======= build bin ==========
# use BUILDPLATFORM to cross build
FROM --platform=${TARGETPLATFORM} ${GOLANG_IMAGE} as builder

# docker build var
ARG TARGETOS
ARG TARGETARCH

# custom var
ARG RACE
ARG NOSTRIP
ARG NOOPT
ARG APT_HTTP_PROXY

ARG USE_PROXY_SOURCE
COPY /images/sources-debian12.list /sources.list
RUN [ "$USE_PROXY_SOURCE" != "true" ] && exit 0 ; \
      rm -f /etc/apt/sources.list ; \
      rm -rf /etc/apt/sources.list.d ; \
      cp /sources.list /etc/apt/sources.list

COPY . /src
WORKDIR /src
RUN  [ -z "${APT_HTTP_PROXY}" ] || { echo "use APT_HTTP_PROXY=${APT_HTTP_PROXY}" ; export http_proxy=${APT_HTTP_PROXY} ; } ; \
    make installBuildTool
RUN  make GOARCH=${TARGETARCH}   \
        RACE=${RACE} NOSTRIP=${NOSTRIP} NOOPT=${NOOPT} \
        DESTDIR_BIN=/tmp/install/${TARGETOS}/${TARGETARCH}/bin \
        build_agent_bin
RUN  make GOARCH=${TARGETARCH}   \
        RACE=${RACE} NOSTRIP=${NOSTRIP} NOOPT=${NOOPT} \
        DESTDIR_BIN=/tmp/install/${TARGETOS}/${TARGETARCH}/bin \
        build_inspect_bin

#====== release image =======

FROM --platform=${TARGETPLATFORM} ${BASE_IMAGE}

LABEL maintainer="maintainer@elf-io"

# TARGETOS is an automatic platform ARG enabled by Docker BuildKit.
ARG TARGETOS
# TARGETARCH is an automatic platform ARG enabled by Docker BuildKit.
ARG TARGETARCH
# linux/amd64
ARG TARGETPLATFORM

ARG GIT_COMMIT_VERSION
ENV ENV_GIT_COMMIT_VERSION=${GIT_COMMIT_VERSION}
ARG GIT_COMMIT_TIME
ENV ENV_GIT_COMMIT_TIMESTAMP=${GIT_COMMIT_TIME}
ARG VERSION
ENV ENV_VERSION=${VERSION}

COPY --from=builder /tmp/install/${TARGETOS}/${TARGETARCH}/bin/*   /usr/bin/
COPY --from=gops-cni-builder /out/${TARGETPLATFORM}/bin/gops /bin/gops

#====modify====
CMD ["/usr/bin/agent"]
