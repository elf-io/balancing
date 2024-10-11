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
      serviceName: http-server-v4
      namespace: default
      # podEndpoint;nodeProxy;nodePort
      redirectMode: podEndpoint
      toPorts:
      - port: "80"
        protocol: TCP
        name: p1
      - port: "80"
        protocol: TCP
        name: p2
EOF


cat <<EOF | kubectl apply -f -
apiVersion: balancing.elf.io/v1beta1
kind: BalancingPolicy
metadata:
  name: test-service-nodeproxy
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
      serviceName: http-server-v4
      namespace: default
      # podEndpoint;nodeProxy;nodePort
      redirectMode: nodeProxy
      toPorts:
      - port: "80"
        protocol: TCP
        name: p1
      - port: "80"
        protocol: TCP
        name: p2
EOF


cat <<EOF | kubectl apply -f -
apiVersion: balancing.elf.io/v1beta1
kind: BalancingPolicy
metadata:
  name: test-service-nodeport
  annotations:
     balancing.elf.io/serviceId: "20003"
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
      serviceName: http-server-v4
      namespace: default
      # podEndpoint;nodeProxy;nodePort
      redirectMode: nodePort
      toPorts:
      - port: "80"
        protocol: TCP
        name: p1
      - port: "80"
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
      - ip: "1.1.1.1"
        port: "9080"
        protocol: TCP
        name: p1
      - ip: "1.1.1.2"
        port: "9080"
        protocol: TCP
        name: p1
      - ip: "1.1.2.1"
        port: "9080"
        protocol: TCP
        name: p2
      - ip: "1.1.2.2"
        port: "9080"
        protocol: TCP
        name: p2
EOF


kubectl get balancingpolicies




