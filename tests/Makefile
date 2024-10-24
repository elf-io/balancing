include ../Makefile.defs


.PHONY: init_kind_env
init_kind_env:
	make init_one_kind
	make install_proscope


.PHONY: init_one_kind
init_one_kind: installE2eTools clean
	@echo "================== init vagrant cluster"
	cd ./vagrant ;  \
			KUBECONFIG_PATH=$(E2E_KUBECONFIG_PATH) ./setupCluster.sh on
	@echo "===================== deploy prometheus CRD ========== "
	# https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/example/prometheus-operator-crd/monitoring.coreos.com_servicemonitors.yaml ; } \
	kubectl apply --kubeconfig $(E2E_KUBECONFIG_PATH)  -f ./crd/monitoring.coreos.com_servicemonitors.yaml
	# https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/example/prometheus-operator-crd/monitoring.coreos.com_podmonitors.yaml ; } \
	kubectl apply --kubeconfig $(E2E_KUBECONFIG_PATH) -f ./crd/monitoring.coreos.com_podmonitors.yaml
	# https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/example/prometheus-operator-crd/monitoring.coreos.com_prometheusrules.yaml ; } \
	kubectl apply --kubeconfig $(E2E_KUBECONFIG_PATH) -f ./crd/monitoring.coreos.com_prometheusrules.yaml
	# https://raw.githubusercontent.com/prometheus-operator/prometheus-operator/main/example/prometheus-operator-crd/monitoring.coreos.com_probes.yaml  ; } \
	kubectl apply --kubeconfig $(E2E_KUBECONFIG_PATH) -f ./crd/monitoring.coreos.com_probes.yaml
	# https://raw.githubusercontent.com/grafana-operator/grafana-operator/master/deploy/manifests/latest/crds.yaml  ; } \
	kubectl apply --kubeconfig $(E2E_KUBECONFIG_PATH) -f ./crd/grafanadashboards.yaml
	@echo "========================================================"
	@echo "   export KUBECONFIG=$(E2E_KUBECONFIG_PATH)             "
	@echo "   kubectl get pod -o wide -A                           "
	@echo "========================================================"


.PHONY: installE2eTools
installE2eTools:
	$(ROOT_DIR)/tests/scripts/installE2eTools.sh

.PHONY: install_proscope
install_proscope:
	if [ -n "$(PYROSCOPE_LOCAL_PORT)" ] ; then \
  		echo "install proscope " ; \
		docker stop $(PYROSCOPE_CONTAINER_NAME) &>/dev/null || true ; \
		docker rm $(PYROSCOPE_CONTAINER_NAME) &>/dev/null || true ; \
		ServerAddress=$$(docker network inspect kind -f {{\(index\ $$.IPAM.Config\ 0\).Gateway}}) ; \
		echo "setup pyroscope on $${ServerAddress}:$(PYROSCOPE_LOCAL_PORT)" ; \
		docker run -d --name $(PYROSCOPE_CONTAINER_NAME) -p $(PYROSCOPE_LOCAL_PORT):4040 $(PYROSCOPE_IMAGE_NAME) server ; \
		echo "finish setuping pyroscope " ; \
      fi



#==================

.PHONY: check_test_app_images_ready
check_test_app_images_ready:
	CHART_LIST=" $(ROOT_DIR)/test/charts/backendserver $(ROOT_DIR)/test/charts/proxyserver $(ROOT_DIR)/test/charts/redirectserver " ;\
		for CHART_DIR in $${CHART_LIST} ; do \
			echo "check image for $${CHART_DIR}  " ; \
			IMAGE_LIST=` helm template test $${CHART_DIR} --set image.tag=$(TEST_APP_IMAGE_TAG)  | grep " image: " | tr -d '"'| awk '{print $$2}' ` ; \
			if [ -z "$${IMAGE_LIST}" ] ; then \
				echo "warning, failed to find image from chart $${CHART_DIR}" ; \
				exit 1 ;\
			else \
				echo "find image from chart : $${IMAGE_LIST} " ; \
				for IMAGE in $${IMAGE_LIST} ; do \
					echo "try to find image $${IMAGE} " ; \
					EXIST=` docker images | awk '{printf("%s:%s\n",$$1,$$2)}' | grep "$${IMAGE}" ` || true ; \
					if [ -z "$${EXIST}" ] ; then \
							CI_IMAGE=$${IMAGE##*/} ; \
							echo "try to find github CI image $${CI_IMAGE} " ; \
							EXIST=` docker images | awk '{printf("%s:%s\n",$$1,$$2)}' | grep "$${CI_IMAGE}" ` || true ; \
							if [ -z "$${EXIST}" ] ; then \
								echo "error, failed to find image $${IMAGE}" ; \
								echo "error, failed to find image $${CI_IMAGE}" ; \
								echo "try to build image 'build_local_test_app_image' " ; \
								exit 1 ; \
							fi ; \
							docker tag $${CI_IMAGE} $${IMAGE} ; \
					fi ;\
					echo "image exists: $${IMAGE}" ; \
				done ; \
				docker images ; \
			fi \
		done

# this will auto tag github ci image : agent:xxx -> github.com/elf-io/balancing/agent:xxx
.PHONY: check_images_ready
check_images_ready:
	echo "check image  " ; \
	IMAGE_LIST=` helm template test $(ROOT_DIR)/charts --set global.imageTagOverride=$(PROJECT_IMAGE_VERSION)  | grep " image: " | tr -d '"'| awk '{print $$2}' ` ; \
	if [ -z "$${IMAGE_LIST}" ] ; then \
		echo "warning, failed to find image from chart " ; \
		exit 1 ;\
	else \
		echo "find image from chart : $${IMAGE_LIST} " ; \
		for IMAGE in $${IMAGE_LIST} ; do \
		  	echo "try to find image $${IMAGE} " ; \
			EXIST=` docker images | awk '{printf("%s:%s\n",$$1,$$2)}' | grep "$${IMAGE}" ` || true ; \
			if [ -z "$${EXIST}" ] ; then \
					CI_IMAGE=$${IMAGE##*/} ; \
			  		echo "try to find github CI image $${CI_IMAGE} " ; \
			  		EXIST=` docker images | awk '{printf("%s:%s\n",$$1,$$2)}' | grep "$${CI_IMAGE}" ` || true ; \
			  		if [ -z "$${EXIST}" ] ; then \
			  			echo "error, failed to find image $${IMAGE}" ; \
			  			echo "error, failed to find image $${CI_IMAGE}" ; \
			  			exit 1 ; \
			  		fi ; \
			  		docker tag $${CI_IMAGE} $${IMAGE} ; \
			fi ;\
			echo "image exists: $${IMAGE}" ; \
		done ; \
		docker images ; \
	fi


# install spidernet on global cluster
.PHONY: deploy_project
deploy_project: E2E_KUBECONFIG_PATH ?= $(E2E_E2E_KUBECONFIG_PATH_PATH)
deploy_project: KIND_CLUSTER_NAME ?= $(E2E_KIND_CLUSTER_NAME)
deploy_project:
	echo "try to load local image tag $(PROJECT_IMAGE_VERSION) " ; \
		IMAGE_LIST=` helm template test $(ROOT_DIR)/charts --set global.imageTagOverride=$(PROJECT_IMAGE_VERSION)  | grep " image: " | tr -d '"'| awk '{print $$2}' ` ; \
		if [ -z "$${IMAGE_LIST}" ] ; then \
			echo "warning, failed to find image from chart " ; \
		else \
			echo "found image from chart : $${IMAGE_LIST} " ; \
			for IMAGE in $${IMAGE_LIST} ; do \
				EXIST=` docker images | awk '{printf("%s:%s\n",$$1,$$2)}' | grep "$${IMAGE}" ` ; \
				if [ -z "$${EXIST}" ] ; then \
				  echo "docker pull $${IMAGE} to local" ; \
				  docker pull $${IMAGE} ; \
				fi ;\
				echo "load local image $${IMAGE} " ; \
				kind load docker-image $${IMAGE}  --name $(KIND_CLUSTER_NAME)  ; \
				if grep "agent" <<< "$${IMAGE}" &>/dev/null ; then \
						AGENT_IMAGE=$${IMAGE} ; \
				fi ; \
			done ; \
		fi  ; \
		echo "begin to deploy on kubernetes" ; \
		helm --kubeconfig=$(E2E_KUBECONFIG_PATH) uninstall -n $(E2E_INSTALL_NAMESPACE) project || true  ;\
		HELM_OPTION="" ; \
    	if [ -n "$(PYROSCOPE_LOCAL_PORT)" ] ; then \
			echo "add env" ; \
			ServerAddress=$$(docker network inspect kind -f {{\(index\ $$.IPAM.Config\ 0\).Gateway}}) ; \
			HELM_OPTION+=" --set balancingAgent.extraEnv[0].name=ENV_PYROSCOPE_PUSH_SERVER_ADDRESS  --set balancingAgent.extraEnv[0].value=http://$${ServerAddress}:$(PYROSCOPE_LOCAL_PORT) " ; \
			HELM_OPTION+=" --set balancingController.extraEnv[0].name=ENV_PYROSCOPE_PUSH_SERVER_ADDRESS  --set balancingController.extraEnv[0].value=http://$${ServerAddress}:$(PYROSCOPE_LOCAL_PORT) " ; \
		fi ; \
		HELM_OPTION+=" --set balancingAgent.debug.logLevel=debug --set balancingController.debug.logLevel=debug " ; \
		HELM_OPTION+=" --set balancingAgent.prometheus.enabled=true --set balancingController.prometheus.enabled=true  " ; \
		HELM_OPTION+=" --set balancingAgent.nodeEntryInterface=eth0 " ; \
		helm --kubeconfig=$(E2E_KUBECONFIG_PATH) install project $(ROOT_DIR)/charts \
				-n $(E2E_INSTALL_NAMESPACE) --create-namespace --wait --debug \
				--set global.imageTagOverride=$(PROJECT_IMAGE_VERSION) \
				$${HELM_OPTION} \
				|| { KIND_CLUSTER_NAME=$(KIND_CLUSTER_NAME) ./scripts/debugCluster.sh $(E2E_KUBECONFIG_PATH) "detail"  $(E2E_INSTALL_NAMESPACE) ; exit 1 ; } ; \
		echo "begin to deploy agent on host node " ; \
		./scripts/runHostNode.sh "runAgent" "$${AGENT_IMAGE}"  ; \
		exit 0

# test kind is ok
.PHONY: install_example_app
install_example_app: E2E_KUBECONFIG_PATH ?= $(E2E_E2E_KUBECONFIG_PATH_PATH)
install_example_app: KIND_CLUSTER_NAME ?= $(E2E_KIND_CLUSTER_NAME)
install_example_app:
	@echo "---------- install applicatio on kubernetes"
	echo "get images ... " ; \
		CHART_LIST=" $(ROOT_DIR)/test/charts/backendserver $(ROOT_DIR)/test/charts/proxyserver $(ROOT_DIR)/test/charts/redirectserver " ;\
		for CHART_DIR in $${CHART_LIST} ; do \
			echo "check image for $${CHART_DIR}  " ; \
			IMAGE_LIST=` helm template test $${CHART_DIR} --set image.tag=$(TEST_APP_IMAGE_TAG)  | grep " image: " | tr -d '"'| awk '{print $$2}' ` ; \
			if [ -z "$${IMAGE_LIST}" ] ; then \
				echo "warning, failed to find image from chart $${CHART_DIR}" ; \
				exit 1 ;\
			else \
				echo "find image from chart : $${IMAGE_LIST} " ; \
				for IMAGE in $${IMAGE_LIST} ; do \
					echo "try to find image $${IMAGE} " ; \
					EXIST=` docker images | awk '{printf("%s:%s\n",$$1,$$2)}' | grep "$${IMAGE}" ` || true ; \
					if [ -z "$${EXIST}" ] ; then \
						  echo "docker pull $${IMAGE} to local" ; \
						  docker pull $${IMAGE} ; \
					fi ;\
					echo "load local image $${IMAGE} " ; \
					kind load docker-image $${IMAGE}  --name $(KIND_CLUSTER_NAME)  ; \
				done ; \
			fi \
		done
	echo "deploy proxy server " ; \
			helm  --kubeconfig=$(E2E_KUBECONFIG_PATH) uninstall  proxyserver &>/dev/null || true ; \
			helm  --kubeconfig=$(E2E_KUBECONFIG_PATH) install  proxyserver ./charts/proxyserver \
					--wait --debug --set image.tag=$(TEST_APP_IMAGE_TAG) \
			|| { KIND_CLUSTER_NAME=$(KIND_CLUSTER_NAME) ./scripts/debugCluster.sh $(E2E_KUBECONFIG_PATH) "detail"  $(E2E_INSTALL_NAMESPACE) ; exit 1 ; } ;\
			echo " deploy proxy on host node" ; \
			IMAGE=` helm template test ./charts/proxyserver --set image.tag=$(TEST_APP_IMAGE_TAG)  | grep " image: " | tr -d '"'| awk '{print $$2}' ` ; \
			./scripts/runHostNode.sh "runProxyServer" "$${IMAGE}"
	echo "deploy backend server " ; \
			helm  --kubeconfig=$(E2E_KUBECONFIG_PATH) uninstall  backendserver &>/dev/null || true ; \
			helm  --kubeconfig=$(E2E_KUBECONFIG_PATH) install  backendserver ./charts/backendserver \
			--wait --debug --set image.tag=$(TEST_APP_IMAGE_TAG) \
			|| { KIND_CLUSTER_NAME=$(KIND_CLUSTER_NAME) ./scripts/debugCluster.sh $(E2E_KUBECONFIG_PATH) "detail"  $(E2E_INSTALL_NAMESPACE) ; exit 1 ; }
	echo "deploy redirect server " ; \
			helm  --kubeconfig=$(E2E_KUBECONFIG_PATH) uninstall  redirectserver &>/dev/null || true ; \
			helm  --kubeconfig=$(E2E_KUBECONFIG_PATH) install  redirectserver ./charts/redirectserver \
			--wait --debug --set image.tag=$(TEST_APP_IMAGE_TAG) \
			|| { KIND_CLUSTER_NAME=$(KIND_CLUSTER_NAME) ./scripts/debugCluster.sh $(E2E_KUBECONFIG_PATH) "detail"  $(E2E_INSTALL_NAMESPACE) ; exit 1 ; }
	@echo "---------- install proxy-server application on host"
	./scripts/testAppConnectivity.sh "$(E2E_KUBECONFIG_PATH)" || { echo "failed to test the connectivity" ; exit 1 ; }
	@echo "========================================================"
	@echo "   deploy kind cluster $(KIND_CLUSTER_NAME)             "
	@echo "   export KUBECONFIG=$(E2E_KUBECONFIG_PATH)                 "
	@echo "   kubectl get pod -o wide -A                           "
	@echo "========================================================"
	@ KUBECONFIG=$(E2E_KUBECONFIG_PATH)  kubectl get pod -o wide -A


.PHONY: clean
clean:
	-@ cd ./vagrant ; ./setupCluster.sh off
	-@ rm -rf $(E2E_RUNTIME_DIR)



#============ e2e ====================
.PHONY: e2e_test
e2e_test: KIND_CLUSTER_NAME ?= $(E2E_KIND_CLUSTER_NAME)
e2e_test: E2E_KUBECONFIG_PATH ?= $(E2E_E2E_KUBECONFIG_PATH_PATH)
e2e_test:
	@echo -e "\033[35m Run e2e test on the cluster $(KIND_CLUSTER_NAME) \033[0m "
	@ echo -e "\033[35m [E2E] Run E2E with ginkgo label=$(E2E_GINKGO_LABELS) , timeout=$(E2E_TIMEOUT) GINKGO_OPTION=$(E2E_GINKGO_OPTION) \033[0m"
	@  NODE_LIST=` docker ps | egrep " kindest/node.* $(KIND_CLUSTER_NAME)-(control|worker)" | awk '{print $$NF }' ` ; \
		[ -n "$$NODE_LIST" ] || { echo "error, failed to find any kind nodes, please setup kind cluster $(KIND_CLUSTER_NAME) first" ; exit 1 ; } ; \
		NODE_LIST=` echo "$${NODE_LIST}" | tr -d ' ' | tr '\n' ',' ` ; \
		NODE_LIST=$${NODE_LIST%%,} ; \
		echo "find cluster node: $${NODE_LIST}" ; \
		export E2E_KIND_CLUSTER_NODE_LIST="$${NODE_LIST}" ; \
		export E2E_CLUSTER_NAME=$(KIND_CLUSTER_NAME) ; \
		if [ "$(E2E_IP_FAMILY)" == "ipv4" ] ; then \
			export E2E_IPV4_ENABLED=true ; export E2E_IPV6_ENABLED=false ; \
		elif [ "$(E2E_IP_FAMILY)" == "ipv6" ] ; then \
			export E2E_IPV4_ENABLED=false ; export E2E_IPV6_ENABLED=true ; \
		else \
			export E2E_IPV4_ENABLED=true ; export E2E_IPV6_ENABLED=true ; \
		fi ; \
		export E2E_KUBECONFIG_PATH=$(E2E_KUBECONFIG_PATH) ; [ -f "$(E2E_KUBECONFIG_PATH)" ] || { echo "error, does not exist KUBECONFIG $(E2E_KUBECONFIG)" ; exit 1 ; } ; \
		rm -f $(E2E_LOG_FILE) || true ; \
		echo "=========== before test `date` ===========" >> $(E2E_LOG_FILE) ; \
		./scripts/debugCluster.sh $(E2E_KUBECONFIG_PATH) "system" "$(E2E_LOG_FILE)" $(E2E_INSTALL_NAMESPACE) ; \
		RESULT=0 ; \
		$(ROOT_DIR)/tools/golang/ginkgo.sh \
			--race --timeout=$(E2E_TIMEOUT) --output-interceptor-mode=none  --slow-spec-threshold=15s \
			--json-report e2ereport.json --output-dir $(ROOT_DIR) --procs $(E2E_GINKGO_PROCS) \
			--label-filter="$(E2E_GINKGO_LABELS)" -randomize-suites -randomize-all  -vv --fail-fast  $(E2E_GINKGO_OPTION) \
			-r e2e/*  || RESULT=1  ; \
		echo "=========== after test `date` ===========" >> $(E2E_LOG_FILE) ; \
		./scripts/debugCluster.sh $(E2E_KUBECONFIG_PATH) "system" "$(E2E_LOG_FILE)" $(E2E_INSTALL_NAMESPACE) ; \
		KIND_CLUSTER_NAME=$(KIND_CLUSTER_NAME) ./scripts/debugCluster.sh $(E2E_KUBECONFIG_PATH) "detail" "$(E2E_LOG_FILE)" $(E2E_INSTALL_NAMESPACE) ; \
		./scripts/debugCluster.sh $(E2E_KUBECONFIG_PATH) "error" "$(E2E_LOG_FILE)" $(E2E_INSTALL_NAMESPACE) || { echo "error, found error log, datarace/pacni/longlock !!!" ; RESULT=1 ; } ; \
		if (($${RESULT} != 0)) ; then \
		   echo "failed to run e2e test"  ; \
		   exit 1 ; \
		fi ; \
		echo "" ; \
		echo "============================================" ; \
		echo "succeeded to run all test" ; \
		echo "output report to e2ereport.json" ; \
		echo "output env log to $(E2E_LOG_FILE) "
