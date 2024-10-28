#!/bin/bash

NAME=http-server
NAMESPACE=default
IMAGE=localhost/weizhoulan/rdmatool:latest

cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: ${NAME}
  namespace: ${NAMESPACE}
  labels:
    app: $NAME
spec:
  selector:
    matchLabels:
      app: $NAME
  template:
    metadata:
      name: $NAME
      labels:
        app: $NAME
    spec:
      containers:
      - name: $NAME
        image: $IMAGE
        imagePullPolicy: IfNotPresent
        command: ["/usr/bin/agent"]
        args: ["--app-mode=true"]
        securityContext:
          privileged: true
        ports:
        - containerPort: 80
          name: http
---
apiVersion: v1
kind: Service
metadata:
  name: $NAME-v4
  namespace: ${NAMESPACE}
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 80
    name: http1
  - port: 8080
    targetPort: 80
    name: http2
  selector:
    app: $NAME
---
apiVersion: v1
kind: Service
metadata:
  name: $NAME-balancing-pod-v4
  namespace: ${NAMESPACE}
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 80
    name: http1
  - port: 8080
    targetPort: 80
    name: http2
  selector:
    app: $NAME
---
apiVersion: v1
kind: Service
metadata:
  name: $NAME-balancing-hostport-v4
  namespace: ${NAMESPACE}
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 80
    name: http1
  - port: 8080
    targetPort: 80
    name: http2
  selector:
    app: $NAME
---
apiVersion: v1
kind: Service
metadata:
  name: $NAME-balancing-nodeproxy-v4
  namespace: ${NAMESPACE}
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 80
    name: http1
  - port: 8080
    targetPort: 80
    name: http2
  selector:
    app: $NAME
---
apiVersion: v1
kind: Service
metadata:
  name: $NAME-affinity-v4
  namespace: ${NAMESPACE}
spec:
  type: NodePort
  sessionAffinity: ClientIP
  sessionAffinityConfig:
    clientIP:
      timeoutSeconds: 60
  ports:
  - port: 80
    targetPort: 80
    name: http
  selector:
    app: $NAME
---
apiVersion: v1
kind: Service
metadata:
  name: $NAME-local-v4
  namespace: ${NAMESPACE}
spec:
  type: NodePort
  externalTrafficPolicy: Local
  internalTrafficPolicy: Local
  ports:
  - port: 80
    targetPort: 80
    name: http
  selector:
    app: $NAME
---
apiVersion: v1
kind: Service
metadata:
  name: $NAME-local-affinity-v4
  namespace: ${NAMESPACE}
spec:
  type: NodePort
  externalTrafficPolicy: Local
  internalTrafficPolicy: Local
  ports:
  - port: 80
    targetPort: 80
    name: http
  selector:
    app: $NAME
---
apiVersion: v1
kind: Service
metadata:
  name: $NAME-external-v4
  namespace: ${NAMESPACE}
spec:
  type: NodePort
  externalIPs:
  - 192.168.255.250
  ports:
  - port: 80
    targetPort: 80
    name: http
  selector:
    app: $NAME
---
apiVersion: v1
kind: Service
metadata:
  name: $NAME-v6
  namespace: ${NAMESPACE}
spec:
  type: LoadBalancer
  ipFamilyPolicy: SingleStack
  ipFamilies:
  - IPv6
  ports:
  - port: 80
    targetPort: 80
    name: http
  selector:
    app: $NAME
EOF



NAME=http-client
cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: ${NAME}
  namespace: ${NAMESPACE}
  labels:
    app: $NAME
spec:
  selector:
    matchLabels:
      app: $NAME
  template:
    metadata:
      name: $NAME
      labels:
        app: $NAME
    spec:
      containers:
      - name: $NAME
        image: $IMAGE
        imagePullPolicy: IfNotPresent
        command: ["/usr/bin/agent"]
        args: ["--app-mode=true"]
        securityContext:
          privileged: true
        ports:
        - containerPort: 80
          name: http
EOF



NAME=http-redirect
cat <<EOF | kubectl apply -f -
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: ${NAME}
  namespace: ${NAMESPACE}
  labels:
    app: $NAME
spec:
  selector:
    matchLabels:
      app: $NAME
  template:
    metadata:
      name: $NAME
      labels:
        app: $NAME
    spec:
      containers:
      - name: $NAME
        image: $IMAGE
        imagePullPolicy: IfNotPresent
        command: ["/usr/bin/agent"]
        args: ["--app-mode=true"]
        securityContext:
          privileged: true
        ports:
        - containerPort: 80
           # 对于 hostPort 部署的应用，同一个 node 上只会有一个 pod 启动成功，多余的 pod 会因为端口占用而启动失败
          hostPort: 20080
          name: http
---
apiVersion: v1
kind: Service
metadata:
  name: $NAME-v4
  namespace: ${NAMESPACE}
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 80
    name: http1
  - port: 8080
    targetPort: 80
    name: http2
  selector:
    app: $NAME
---
apiVersion: v1
kind: Service
metadata:
  name: $NAME-v6
  namespace: ${NAMESPACE}
spec:
  type: LoadBalancer
  ipFamilyPolicy: SingleStack
  ipFamilies:
  - IPv6
  ports:
  - port: 80
    targetPort: 80
    name: http
  selector:
    app: $NAME
EOF
