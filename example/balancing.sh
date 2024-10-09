# balancingpolicy: redirect the request to the endpoint in the cluster

cat <<EOF | kubectl apply -f -
apiVersion: balancing.elf.io/v1beta1
kind: BalancingPolicy
metadata:
  name: test-service-podendpoint
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
        port: 9080
        protocol: TCP
        name: p1
      - ip: "1.1.1.2"
        port: 9080
        protocol: TCP
        name: p1
      - ip: "1.1.2.1"
        port: 9080
        protocol: TCP
        name: p2
      - ip: "1.1.2.2"
        port: 9080
        protocol: TCP
        name: p2
EOF


kubectl get balancingpolicy




