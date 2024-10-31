# develop

## local develop

1. ` make build_local_image  -e USE_PROXY_SOURCE=true  `
   ` make build_local_image  -e APT_HTTP_PROXY=http://10.64.0.3:7890 `

2. ` make build_local_test_app_image   `

3. 
```
make e2e_init  -e E2E_SKIP_KUBE_PROXY=true
```

4.

```
make e2e_deploy
#or
make e2e_deploy -e PROJECT_IMAGE_TAG=8877a79da7c0a9f159363660b5b23e5458480aea \
                -e TEST_APP_IMAGE_TAG=aa7693a44e205c13e9bd3bee63260c9c1048ce24
```

5. 

```shell
make e2e_test_connectivity

```

6. check proscope, browser visits http://NodeIP:28000

7. check metric
