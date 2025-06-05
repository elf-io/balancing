# 搭建开发环境

## 主机软件准备

* Vvagrant 和 VirtualBox
* helm
* kubectl
* jq

## 基于 Vagrant 虚拟机搭建本地开发环境

1. 构建 balancing 镜像

    ```shell
    make build_local_image

    # 对于中国区用户，可以使用代理源来加速构建
    make build_local_image -e USE_PROXY_SOURCE=true
    ```

2. 构建测试应用镜像

    ```shell
    make build_local_test_app_image
    ```

3. 部署基于虚拟机的双节点 Kubernetes 集群（不安装 kube-proxy 组件）

    ```shell
    make e2e_init -e E2E_SKIP_KUBE_PROXY=true
    ```

4. 部署 balancing 和测试应用到集群中

    ```shell
    make e2e_deploy
    
    # 或者使用指定的镜像标签
    make e2e_deploy -e PROJECT_IMAGE_TAG=8877a79da7c0a9f159363660b5b23e5458480aea \
                    -e TEST_APP_IMAGE_TAG=aa7693a44e205c13e9bd3bee63260c9c1048ce24
    ```

5. 测试 balancing 的各种策略例子

    ```shell
    make e2e_test_connectivity
    ```

6. 使用浏览器访问 `http://NodeIP:28000`，查看 Proscope Server 中的 Golang 采样数据。

## 本机调试 ebpf

```
apt-get install clang llvm gcc-multilib libbpf-dev linux-headers-$(uname -r)

cd pkg/ebpf/bpf

# 检测语法错误
clang -fsyntax-only -I.  cgroup.c

```