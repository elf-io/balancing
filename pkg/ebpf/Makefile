#!/usr/bin/make -f

all: install

.PHONY: install
install: buildGolang
	./main

.PHONY: buildGolang
buildGolang:
	rm bpf_*.go || true
	export GOPROXY="https://goproxy.io|https://goproxy.cn|direct" && \
		go mod tidy && \
		go generate ./...
	# To further reduce runtime dependencies, add CGO_ENABLED=0 to go build and your application won't depend on libc
	CGO_ENABLED=0 go build


.PHONY: show_log
show_log:
	bpftool prog tracelog


.PHONY: show_map
show_map:
	bpftool map dump name map_floatip_v4
	bpftool map dump name map_backend_v4


.PHONY: show_prog
show_prog:
	bpftool cgroup tree /run/welan/cgroupv2
	bpftool prog show -f


.PHONY: tool
tool:
	apt-get update && apt-get install -y clang llvm gcc-multilib libbpf-dev linux-headers-$(uname -r) golang


#======================

CGROUP_PATH := /run/welan/cgroupv2
PROG_PINNED_PATH := /sys/fs/bpf/welan

.PHONY: installRaw
installRaw: uninstallRaw build
	mount -t bpf none /sys/fs/bpf/
	mkdir -p $(CGROUP_PATH) || true
	mount -t cgroup2 none $(CGROUP_PATH) || true
	bpftool prog loadall bpf/cgroup.o  $(PROG_PINNED_PATH)
	bpftool cgroup attach $(CGROUP_PATH) connect4  pinned $(PROG_PINNED_PATH)/sock4_connect multi
	bpftool cgroup attach $(CGROUP_PATH) sendmsg4  pinned $(PROG_PINNED_PATH)/sock4_sendmsg multi
	bpftool cgroup attach $(CGROUP_PATH) recvmsg4  pinned $(PROG_PINNED_PATH)/sock4_recvmsg multi
	bpftool cgroup attach $(CGROUP_PATH) getpeername4  pinned $(PROG_PINNED_PATH)/sock4_getpeername multi


.PHONY: buildRaw
buildRaw:
	rm -f bpf/cgroup.o || true
	bpftool btf dump file /sys/kernel/btf/vmlinux format c > ./bpf/vmlinux.h
	clang -O2 -g -target bpf -emit-llvm -c bpf/cgroup.c -o bpf/cgroup.o
	llc -march=bpf -mcpu=probe -mattr=dwarfris -filetype=obj -o bpf/cgroup.o  bpf/cgroup.o

.PHONY: uninstallRaw
uninstallRaw:
	bpftool cgroup detach $(CGROUP_PATH) connect4 name sock4_connect || true
	bpftool cgroup detach $(CGROUP_PATH) sendmsg4 name sock4_sendmsg || true
	bpftool cgroup detach $(CGROUP_PATH) recvmsg4 name sock4_recvmsg || true
	bpftool cgroup detach $(CGROUP_PATH) getpeername4 name sock4_getpeername || true
	rm $(PROG_PINNED_PATH)/* -rf || true
	rm $(PROG_PINNED_PATH) -rf || true
	umount $(CGROUP_PATH) || true
	rm -rf $(CGROUP_PATH) || true
