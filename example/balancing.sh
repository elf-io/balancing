# balancingpolicy: redirect the request to the endpoint in the cluster

cat <<EOF | kubectl apply -f -
apiVersion: balancing.elf.io/v1beta1
kind: balancingpolicy
metadata:
  name: test-service-podEndpoint
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
kind: balancingpolicy
metadata:
  name: test-service-nodeProxy
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
kind: balancingpolicy
metadata:
  name: test-service-nodePort
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
kind: balancingpolicy
metadata:
  name: test-service-nodeProxy
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
kind: localredirectpolicy
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
        toPorts: 9080
        name: p1
      - ip: "1.1.1.2"
        toPorts: 9080
        name: p1
      - ip: "1.1.2.1"
        toPorts: 9081
        name: p2
      - ip: "1.1.2.2"
        toPorts: 9081
        name: p2
EOF


kubectl get balancingpolicy




