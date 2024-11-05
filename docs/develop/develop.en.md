# Setting Up Development Environment

## Host Software Preparation

* vagrant and VirtualBox
* helm
* kubectl
* jq

## Setting Up Local Development Environment with Vagrant VM

1. Build the balancing image

    ```shell
    make build_local_image
    ```

    > For users in China, you can use a proxy source to speed up the build:
    > `make build_local_image -e USE_PROXY_SOURCE=true`

2. Build the test application image

    ```shell
    make build_local_test_app_image
    ```

3. Deploy a dual-node Kubernetes cluster based on VMs (without kube-proxy component)

    ```shell
    make e2e_init -e E2E_SKIP_KUBE_PROXY=true
    ```

4. Deploy balancing and test applications to the cluster

    ```shell
    make e2e_deploy
    # Or use specified image tags
    make e2e_deploy -e PROJECT_IMAGE_TAG=8877a79da7c0a9f159363660b5b23e5458480aea \
                    -e TEST_APP_IMAGE_TAG=aa7693a44e205c13e9bd3bee63260c9c1048ce24
    ```

5. Test various strategies of balancing

    ```shell
    make e2e_test_connectivity
    ```

6. Access `http://NodeIP:28000` in a browser to observe Golang sampling data in the Proscope Server.
