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
      # controller 进行限制，只能有一个 BalancingPolicy 绑定 同名 service  , 否则 agent 侧会 因为 clusterIP 相同 相互覆盖数据
      serviceName: http-server-balacning-pod-v4
      namespace: default
      toPorts:
        # the port and protocol must be in line with the service , but the name should not be
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
      # for podEndpoint: the destination IP is podIP, the destination port is pod port
      redirectMode: podEndpoint
      toPorts:
          # 只能有一个 name: p1
          # port is podPort for podEndpoint mode
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
      # controller 进行限制，只能有一个 BalancingPolicy 绑定 同名 service   , 否则 agent 侧会 相互覆盖数据
      serviceName: http-server-balancing-hostport-v4
      namespace: default
      toPorts:
        # the port and protocol must be in line with the service , but the name should not be
        - port: "80"
          protocol: TCP
          name: p2
  backend:
    serviceEndpoint:
      endpointSelector:
        matchLabels:
          app: http-redirect
      # for hostPort: the destination IP is node IP, the destination port is hostPort
      # 对于 hostPort 部署的应用，同一个 node 上只会有一个 pod 启动成功，多余的 pod 会因为端口占用而启动失败
      redirectMode: hostPort
      toPorts:
          # port is hostPort for hostPort mode
        - port: "20080"
          protocol: TCP
          name: p2
EOF



cat <<EOF | kubectl apply -f -
apiVersion: balancing.elf.io/v1beta1
kind: BalancingPolicy
metadata:
  name: test-service-nodeproxy
  annotations:
     balancing.elf.io/serviceId: "20003"
spec:
  enabled: true
  frontend:
    serviceMatcher:
      # controller 进行限制，只能有一个 BalancingPolicy 绑定 同名 service   , 否则 agent 侧会 相互覆盖数据
      serviceName: http-server-balancing-nodeproxy-v4
      namespace: default
      toPorts:
        # the port and protocol must be in line with the service , but the name should not be
        - port: "80"
          protocol: TCP
          name: p2
  backend:
    serviceEndpoint:
      endpointSelector:
        matchLabels:
          app: http-redirect
      # for hostPort: the destination IP is node porxy IP(tunnel ip), the destination port is nodeProxyPort taken effect by agent
      redirectMode: nodeProxy
      toPorts:
          # port is nodeProxyPort for nodeProxy mode
        - port: "20080"
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
      # controller 进行限制，只能有一个 BalancingPolicy 绑定 相同 ip   , 否则 agent 侧会 相互覆盖数据
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




