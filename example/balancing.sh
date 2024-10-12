# balancingpolicy: redirect the request to the endpoint in the cluster

kubectl get balancingpolicies  | awk '{print $1}' | sed '1 d' | xargs -n 1 -i kubectl delete balancingpolicies {}


cat <<EOF | kubectl apply -f -
apiVersion: balancing.elf.io/v1beta1
kind: BalancingPolicy
metadata:
  name: test-service-podendpoint
  annotations:
     balancing.elf.io/serviceId: "20001"
spec:
  enabled: true
  frontend:
    serviceMatcher:
      serviceName: http-server-v4
      namespace: default
      toPorts:
        - port: "8080"
          protocol: TCP
          name: p1
        - port: "80"
          protocol: TCP
          name: p2
  backend:
    serviceEndpoint:
      endpointSelector:
        matchLabels:
          app: http-redirect
      redirectMode: podEndpoint
      toPorts:
          # 只能有一个 name: p1
        - port: "80"
          protocol: TCP
          name: p1
          # 只能有一个 name: p2
        - port: "80"
          protocol: TCP
          name: p2
EOF


cat <<EOF | kubectl apply -f -
apiVersion: balancing.elf.io/v1beta1
kind: BalancingPolicy
metadata:
  name: test-service-hostport
  annotations:
     balancing.elf.io/serviceId: "20002"
spec:
  enabled: true
  frontend:
    serviceMatcher:
      serviceName: http-server-v4
      namespace: default
      toPorts:
        - port: "8080"
          protocol: TCP
          name: p1
        - port: "80"
          protocol: TCP
          name: p2
  backend:
    serviceEndpoint:
      endpointSelector:
        matchLabels:
          app: http-redirect
      # 对于 hostPort 部署的应用，同一个 node 上只会有一个 pod 启动成功，多余的 pod 会因为端口占用而启动失败
      redirectMode: hostPort
      toPorts:
          # 只能有一个 name: p1
        - hostPort: "20080"
          protocol: TCP
          name: p1
          # 只能有一个 name: p2
        - hostPort: "20080"
          protocol: TCP
          name: p2
EOF



cat <<EOF | kubectl apply -f -
apiVersion: balancing.elf.io/v1beta1
kind: BalancingPolicy
metadata:
  name: test-service-hostport
  annotations:
     balancing.elf.io/serviceId: "20002"
spec:
  enabled: true
  frontend:
    serviceMatcher:
      serviceName: http-server-v4
      namespace: default
      toPorts:
        - port: "8080"
          protocol: TCP
          name: p1
        - port: "80"
          protocol: TCP
          name: p2
  backend:
    serviceEndpoint:
      endpointSelector:
        matchLabels:
          app: http-redirect
      redirectMode: nodeProxy
      toPorts:
          # 只能有一个 name: p1
        - nodeProxyPort: "20080"
          port: "80"
          protocol: TCP
          name: p1
          # 只能有一个 name: p2
        - nodeProxyPort: "20080"
          port: "80"
          protocol: TCP
          name: p2
EOF



cat <<EOF | kubectl apply -f -
apiVersion: balancing.elf.io/v1beta1
kind: BalancingPolicy
metadata:
  name: test-addr
  annotations:
     balancing.elf.io/serviceId: "20004"
spec:
  enabled: true
  frontend:
    addressMatcher:
      ip: "169.254.169.254"
      toPorts:
        - port: "8080"
          protocol: TCP
          name: p1
        - port: "80"
          protocol: TCP
          name: p2
  backend:
    addressEndpoint:
      addresses:
        - "1.1.1.1"
        - "1.1.1.2"
      toPorts:
          # 只能有一个 name: p1
        - port: "80"
          protocol: TCP
          name: p1
          # 只能有一个 name: p2
        - port: "80"
          protocol: TCP
          name: p2
EOF


kubectl get balancingpolicies




