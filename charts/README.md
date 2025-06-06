# balancing

## Introduction

## Features

## Parameters

### Global parameters

| Name                           | Description                                | Value           |
| ------------------------------ | ------------------------------------------ | --------------- |
| `global.imageRegistryOverride` | Global Docker image registry               | `""`            |
| `global.imageTagOverride`      | Global Docker image tag                    | `""`            |
| `global.name`                  | instance name                              | `balancing`     |
| `global.clusterDnsDomain`      | cluster dns domain                         | `cluster.local` |
| `global.commonAnnotations`     | Annotations to add to all deployed objects | `{}`            |
| `global.commonLabels`          | Labels to add to all deployed objects      | `{}`            |
| `global.configName`            | the configmap name                         | `balancing`     |

### feature parameters

| Name                       | Description                                                       | Value   |
| -------------------------- | ----------------------------------------------------------------- | ------- |
| `feature.enableIPv4`       | enable ipv4                                                       | `true`  |
| `feature.enableIPv6`       | enable ipv6                                                       | `false` |
| `feature.redirectQosLimit` | the QoS limit for redirect traffic (requests per second)          | `1`     |
| `feature.apiServerHost`    | the host address of api server, which should not be the clusterIP | `""`    |
| `feature.apiServerPort`    | the host port of api server, which should not be the clusterIP    | `""`    |

### balancingAgent parameters

| Name                                                        | Description                                                                                     | Value                    |
| ----------------------------------------------------------- | ----------------------------------------------------------------------------------------------- | ------------------------ |
| `balancingAgent.name`                                       | the balancingAgent name                                                                         | `balancing-agent`        |
| `balancingAgent.cmdBinName`                                 | the binary name of balancingAgent                                                               | `/usr/bin/agent`         |
| `balancingAgent.hostnetwork`                                | enable hostnetwork mode of balancingAgent pod                                                   | `true`                   |
| `balancingAgent.nodeEntryInterface`                         | set the interface name of each node for entryIP                                                 | `""`                     |
| `balancingAgent.image.registry`                             | the image registry of balancingAgent                                                            | `ghcr.io`                |
| `balancingAgent.image.repository`                           | the image repository of balancingAgent                                                          | `elf-io/balancing-agent` |
| `balancingAgent.image.pullPolicy`                           | the image pullPolicy of balancingAgent                                                          | `IfNotPresent`           |
| `balancingAgent.image.digest`                               | the image digest of balancingAgent, which takes preference over tag                             | `""`                     |
| `balancingAgent.image.tag`                                  | the image tag of balancingAgent, overrides the image tag whose default is the chart appVersion. | `""`                     |
| `balancingAgent.image.imagePullSecrets`                     | the image imagePullSecrets of balancingAgent                                                    | `[]`                     |
| `balancingAgent.serviceAccount.create`                      | create the service account for the balancingAgent                                               | `true`                   |
| `balancingAgent.serviceAccount.annotations`                 | the annotations of balancingAgent service account                                               | `{}`                     |
| `balancingAgent.service.annotations`                        | the annotations for balancingAgent service                                                      | `{}`                     |
| `balancingAgent.service.type`                               | the type for balancingAgent service                                                             | `ClusterIP`              |
| `balancingAgent.priorityClassName`                          | the priority Class Name for balancingAgent                                                      | `system-node-critical`   |
| `balancingAgent.affinity`                                   | the affinity of balancingAgent                                                                  | `{}`                     |
| `balancingAgent.extraArgs`                                  | the additional arguments of balancingAgent container                                            | `[]`                     |
| `balancingAgent.extraEnv`                                   | the additional environment variables of balancingAgent container                                | `[]`                     |
| `balancingAgent.extraVolumes`                               | the additional volumes of balancingAgent container                                              | `[]`                     |
| `balancingAgent.extraVolumeMounts`                          | the additional hostPath mounts of balancingAgent container                                      | `[]`                     |
| `balancingAgent.podAnnotations`                             | the additional annotations of balancingAgent pod                                                | `{}`                     |
| `balancingAgent.podLabels`                                  | the additional label of balancingAgent pod                                                      | `{}`                     |
| `balancingAgent.resources.limits.cpu`                       | the cpu limit of balancingAgent pod                                                             | `1000m`                  |
| `balancingAgent.resources.limits.memory`                    | the memory limit of balancingAgent pod                                                          | `1024Mi`                 |
| `balancingAgent.resources.requests.cpu`                     | the cpu requests of balancingAgent pod                                                          | `100m`                   |
| `balancingAgent.resources.requests.memory`                  | the memory requests of balancingAgent pod                                                       | `128Mi`                  |
| `balancingAgent.securityContext.privileged`                 | the securityContext privileged of balancingAgent daemonset pod                                  | `true`                   |
| `balancingAgent.httpServer.port`                            | the http Port for balancingAgent, for health checking                                           | `5810`                   |
| `balancingAgent.httpServer.startupProbe.failureThreshold`   | the failure threshold of startup probe for balancingAgent health checking                       | `60`                     |
| `balancingAgent.httpServer.startupProbe.periodSeconds`      | the period seconds of startup probe for balancingAgent health checking                          | `2`                      |
| `balancingAgent.httpServer.livenessProbe.failureThreshold`  | the failure threshold of startup probe for balancingAgent health checking                       | `6`                      |
| `balancingAgent.httpServer.livenessProbe.periodSeconds`     | the period seconds of startup probe for balancingAgent health checking                          | `10`                     |
| `balancingAgent.httpServer.readinessProbe.failureThreshold` | the failure threshold of startup probe for balancingAgent health checking                       | `3`                      |
| `balancingAgent.httpServer.readinessProbe.periodSeconds`    | the period seconds of startup probe for balancingAgent health checking                          | `10`                     |
| `balancingAgent.prometheus.enabled`                         | enable template agent to collect metrics                                                        | `false`                  |
| `balancingAgent.prometheus.port`                            | the metrics port of template agent                                                              | `5811`                   |
| `balancingAgent.prometheus.serviceMonitor.install`          | install serviceMonitor for template agent. This requires the prometheus CRDs to be available    | `false`                  |
| `balancingAgent.prometheus.serviceMonitor.namespace`        | the serviceMonitor namespace. Default to the namespace of helm instance                         | `""`                     |
| `balancingAgent.prometheus.serviceMonitor.annotations`      | the additional annotations of balancingAgent serviceMonitor                                     | `{}`                     |
| `balancingAgent.prometheus.serviceMonitor.labels`           | the additional label of balancingAgent serviceMonitor                                           | `{}`                     |
| `balancingAgent.prometheus.prometheusRule.install`          | install prometheusRule for template agent. This requires the prometheus CRDs to be available    | `false`                  |
| `balancingAgent.prometheus.prometheusRule.namespace`        | the prometheusRule namespace. Default to the namespace of helm instance                         | `""`                     |
| `balancingAgent.prometheus.prometheusRule.annotations`      | the additional annotations of balancingAgent prometheusRule                                     | `{}`                     |
| `balancingAgent.prometheus.prometheusRule.labels`           | the additional label of balancingAgent prometheusRule                                           | `{}`                     |
| `balancingAgent.prometheus.grafanaDashboard.install`        | install grafanaDashboard for template agent. This requires the prometheus CRDs to be available  | `false`                  |
| `balancingAgent.prometheus.grafanaDashboard.namespace`      | the grafanaDashboard namespace. Default to the namespace of helm instance                       | `""`                     |
| `balancingAgent.prometheus.grafanaDashboard.annotations`    | the additional annotations of balancingAgent grafanaDashboard                                   | `{}`                     |
| `balancingAgent.prometheus.grafanaDashboard.labels`         | the additional label of balancingAgent grafanaDashboard                                         | `{}`                     |
| `balancingAgent.debug.logLevel`                             | the log level of template agent [debug, info, warn, error, fatal, panic]                        | `info`                   |
| `balancingAgent.debug.gopsPort`                             | the gops port of template agent                                                                 | `5812`                   |

### balancingController parameters

| Name                                                             | Description                                                                                                                    | Value                         |
| ---------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------ | ----------------------------- |
| `balancingController.name`                                       | the balancingController name                                                                                                   | `balancing-controller`        |
| `balancingController.replicas`                                   | the replicas number of balancingController pod                                                                                 | `1`                           |
| `balancingController.cmdBinName`                                 | the binName name of balancingController                                                                                        | `/usr/bin/controller`         |
| `balancingController.hostnetwork`                                | enable hostnetwork mode of balancingController pod. Notice, if no CNI available before template installation, must enable this | `true`                        |
| `balancingController.image.registry`                             | the image registry of balancingController                                                                                      | `ghcr.io`                     |
| `balancingController.image.repository`                           | the image repository of balancingController                                                                                    | `elf-io/balancing-controller` |
| `balancingController.image.pullPolicy`                           | the image pullPolicy of balancingController                                                                                    | `IfNotPresent`                |
| `balancingController.image.digest`                               | the image digest of balancingController, which takes preference over tag                                                       | `""`                          |
| `balancingController.image.tag`                                  | the image tag of balancingController, overrides the image tag whose default is the chart appVersion.                           | `""`                          |
| `balancingController.image.imagePullSecrets`                     | the image imagePullSecrets of balancingController                                                                              | `[]`                          |
| `balancingController.serviceAccount.create`                      | create the service account for the balancingController                                                                         | `true`                        |
| `balancingController.serviceAccount.annotations`                 | the annotations of balancingController service account                                                                         | `{}`                          |
| `balancingController.service.annotations`                        | the annotations for balancingController service                                                                                | `{}`                          |
| `balancingController.service.type`                               | the type for balancingController service                                                                                       | `ClusterIP`                   |
| `balancingController.priorityClassName`                          | the priority Class Name for balancingController                                                                                | `system-node-critical`        |
| `balancingController.affinity`                                   | the affinity of balancingController                                                                                            | `{}`                          |
| `balancingController.extraArgs`                                  | the additional arguments of balancingController container                                                                      | `[]`                          |
| `balancingController.extraEnv`                                   | the additional environment variables of balancingController container                                                          | `[]`                          |
| `balancingController.extraVolumes`                               | the additional volumes of balancingController container                                                                        | `[]`                          |
| `balancingController.extraVolumeMounts`                          | the additional hostPath mounts of balancingController container                                                                | `[]`                          |
| `balancingController.podAnnotations`                             | the additional annotations of balancingController pod                                                                          | `{}`                          |
| `balancingController.podLabels`                                  | the additional label of balancingController pod                                                                                | `{}`                          |
| `balancingController.securityContext`                            | the security Context of balancingController pod                                                                                | `{}`                          |
| `balancingController.resources.limits.cpu`                       | the cpu limit of balancingController pod                                                                                       | `500m`                        |
| `balancingController.resources.limits.memory`                    | the memory limit of balancingController pod                                                                                    | `1024Mi`                      |
| `balancingController.resources.requests.cpu`                     | the cpu requests of balancingController pod                                                                                    | `100m`                        |
| `balancingController.resources.requests.memory`                  | the memory requests of balancingController pod                                                                                 | `128Mi`                       |
| `balancingController.podDisruptionBudget.enabled`                | enable podDisruptionBudget for balancingController pod                                                                         | `false`                       |
| `balancingController.podDisruptionBudget.minAvailable`           | minimum number/percentage of pods that should remain scheduled.                                                                | `1`                           |
| `balancingController.httpServer.port`                            | the http Port for balancingController, for health checking and http service                                                    | `5820`                        |
| `balancingController.httpServer.startupProbe.failureThreshold`   | the failure threshold of startup probe for balancingController health checking                                                 | `30`                          |
| `balancingController.httpServer.startupProbe.periodSeconds`      | the period seconds of startup probe for balancingController health checking                                                    | `2`                           |
| `balancingController.httpServer.livenessProbe.failureThreshold`  | the failure threshold of startup probe for balancingController health checking                                                 | `6`                           |
| `balancingController.httpServer.livenessProbe.periodSeconds`     | the period seconds of startup probe for balancingController health checking                                                    | `10`                          |
| `balancingController.httpServer.readinessProbe.failureThreshold` | the failure threshold of startup probe for balancingController health checking                                                 | `3`                           |
| `balancingController.httpServer.readinessProbe.periodSeconds`    | the period seconds of startup probe for balancingController health checking                                                    | `10`                          |
| `balancingController.webhookPort`                                | the http port for balancingController webhook                                                                                  | `5822`                        |
| `balancingController.prometheus.enabled`                         | enable template Controller to collect metrics                                                                                  | `false`                       |
| `balancingController.prometheus.port`                            | the metrics port of template Controller                                                                                        | `5821`                        |
| `balancingController.prometheus.serviceMonitor.install`          | install serviceMonitor for template agent. This requires the prometheus CRDs to be available                                   | `false`                       |
| `balancingController.prometheus.serviceMonitor.namespace`        | the serviceMonitor namespace. Default to the namespace of helm instance                                                        | `""`                          |
| `balancingController.prometheus.serviceMonitor.annotations`      | the additional annotations of balancingController serviceMonitor                                                               | `{}`                          |
| `balancingController.prometheus.serviceMonitor.labels`           | the additional label of balancingController serviceMonitor                                                                     | `{}`                          |
| `balancingController.prometheus.prometheusRule.install`          | install prometheusRule for template agent. This requires the prometheus CRDs to be available                                   | `false`                       |
| `balancingController.prometheus.prometheusRule.namespace`        | the prometheusRule namespace. Default to the namespace of helm instance                                                        | `""`                          |
| `balancingController.prometheus.prometheusRule.annotations`      | the additional annotations of balancingController prometheusRule                                                               | `{}`                          |
| `balancingController.prometheus.prometheusRule.labels`           | the additional label of balancingController prometheusRule                                                                     | `{}`                          |
| `balancingController.prometheus.grafanaDashboard.install`        | install grafanaDashboard for template agent. This requires the prometheus CRDs to be available                                 | `false`                       |
| `balancingController.prometheus.grafanaDashboard.namespace`      | the grafanaDashboard namespace. Default to the namespace of helm instance                                                      | `""`                          |
| `balancingController.prometheus.grafanaDashboard.annotations`    | the additional annotations of balancingController grafanaDashboard                                                             | `{}`                          |
| `balancingController.prometheus.grafanaDashboard.labels`         | the additional label of balancingController grafanaDashboard                                                                   | `{}`                          |
| `balancingController.debug.logLevel`                             | the log level of template Controller [debug, info, warn, error, fatal, panic]                                                  | `info`                        |
| `balancingController.debug.gopsPort`                             | the gops port of template Controller                                                                                           | `5824`                        |
| `balancingController.tls.method`                                 | the method for generating TLS certificates. [ provided , certmanager , auto]                                                   | `auto`                        |
| `balancingController.tls.secretName`                             | the secret name for storing TLS certificates                                                                                   | `balancing-webhook-certs`     |
| `balancingController.tls.certmanager.certValidityDuration`       | generated certificates validity duration in days for 'certmanager' method                                                      | `365`                         |
| `balancingController.tls.certmanager.issuerName`                 | issuer name of cert manager 'certmanager'. If not specified, a CA issuer will be created.                                      | `""`                          |
| `balancingController.tls.certmanager.extraDnsNames`              | extra DNS names added to certificate when it's auto generated                                                                  | `[]`                          |
| `balancingController.tls.certmanager.extraIPAddresses`           | extra IP addresses added to certificate when it's auto generated                                                               | `[]`                          |
| `balancingController.tls.provided.tlsCert`                       | encoded tls certificate for provided method                                                                                    | `""`                          |
| `balancingController.tls.provided.tlsKey`                        | encoded tls key for provided method                                                                                            | `""`                          |
| `balancingController.tls.provided.tlsCa`                         | encoded tls CA for provided method                                                                                             | `""`                          |
| `balancingController.tls.auto.caExpiration`                      | ca expiration for auto method                                                                                                  | `73000`                       |
| `balancingController.tls.auto.certExpiration`                    | server cert expiration for auto method                                                                                         | `73000`                       |
| `balancingController.tls.auto.extraIpAddresses`                  | extra IP addresses of server certificate for auto method                                                                       | `[]`                          |
| `balancingController.tls.auto.extraDnsNames`                     | extra DNS names of server cert for auto method                                                                                 | `[]`                          |
