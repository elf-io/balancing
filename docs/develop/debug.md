# develop

## local develop

1. ` make build_local_image -e  APT_HTTP_PROXY=http://10.64.0.3:7890 `

2. ` make build_local_test_app_image  -e APT_HTTP_PROXY=http://10.64.0.3:7890 `

3. ` make e2e_init  `

4. 

```
make e2e_deploy
#or
make e2e_deploy -e PROJECT_IMAGE_TAG=228ebcbda632481f9bf7471983d4dab2fc06b74e \
                -e TEST_APP_IMAGE_TAG=5b44647869130c82d9582eedc9b5c553aece80b7
```

5. 

```shell
make e2e_test_connectivity

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
