# develop

## local develop

1. ` make build_local_image -e  APT_HTTP_PROXY=http://10.64.0.3:7890 `

2. ` make build_local_test_app_image  -e APT_HTTP_PROXY=http://10.64.0.3:7890 `

3. ` make e2e_init  `

4. 

```
make e2e_deploy
#or
make e2e_deploy -e PROJECT_IMAGE_TAG=cfae79fa0283efa646afc633a7837290701d6a13 \
                -e TEST_APP_IMAGE_TAG=c357da95d2bbe22e2573753d63e7dbf44d1d2edd
```

6. check proscope, browser visits http://NodeIP:28000

7. check metric

## chart develop

helm repo add rocktemplate https://spidernet-io.github.io/rocktemplate

## test

```shell

cat <<EOF > mybook1.yaml
apiVersion: rocktemplate.spidernet.io/v1
kind: Mybook
metadata:
  name: test1
spec:
  ipVersion: 4
  subnet: "1.0.0.0/8"
EOF

kubectl apply -f mybook1.yaml


cat <<EOF > mybook2.yaml
apiVersion: rocktemplate.spidernet.io/v1
kind: Mybook
metadata:
  name: test2
spec:
  ipVersion: 4
  subnet: "2.0.0.0/8"
EOF

kubectl apply -f mybook2.yaml


```
