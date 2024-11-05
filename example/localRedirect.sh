#!/bin/bash
# Copyright 2024 Authors of elf-io
# SPDX-License-Identifier: Apache-2.0

# localredirectpolicy: redirect the request to the pod in the local node

kubectl get localredirectpolicies  | awk '{print $1}' | sed '1 d' | xargs -n 1 -i kubectl delete localredirectpolicies {}

cat <<EOF | kubectl apply -f -
apiVersion: balancing.elf.io/v1beta1
kind: LocalRedirectPolicy
metadata:
  name: test-service
spec:
  enabled: true
  frontend:
    serviceMatcher:
      # controller 进行限制，只能有一个 LocalRedirectPolicy 绑定 同名 service   , 否则 agent 侧会 相互覆盖数据
      serviceName: http-server-v4
      namespace: default
      toPorts:
        # 只能有一个 name: p1
        - port: "8080"
          protocol: TCP
          name: p1
        # 只能有一个 name: p2
        - port: "80"
          protocol: TCP
          name: p2
  backend:
    endpointSelector:
      matchLabels:
        app: http-redirect
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
kind: LocalRedirectPolicy
metadata:
  name: test-addr
  annotations:
     balancing.elf.io/serviceId: "10091"
spec:
  enabled: true
  frontend:
    addressMatcher:
      ip: "169.254.0.90"
      toPorts:
        # 只能有一个 name: p1
        - port: "8080"
          protocol: TCP
          name: p1
        # 只能有一个 name: p2
        - port: "80"
          protocol: TCP
          name: p2
  backend:
    endpointSelector:
      matchLabels:
        app: http-redirect
    toPorts:
      - port: "80"
        protocol: TCP
        name: p1
      - port: "80"
        protocol: TCP
        name: p2
EOF


kubectl get localredirectpolicies



