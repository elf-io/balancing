# develop

## local develop

1. `make build_local_image -e  APT_HTTP_PROXY=http://10.64.0.14:7890`

2. `make e2e_init`

3. `make e2e_run`

4. check proscope, browser visits http://NodeIP:4040

5. check metric,

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
