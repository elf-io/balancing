include ../Makefile.defs


.PHONY: init_env
init_env: init_vagrant install_proscope showInfo

.PHONY: init_vagrant
init_vagrant: installE2eTools clean
	@echo "================== init vagrant cluster"
	cd ./vagrant ;  \
		  HOSTPORT_CONTROLVM_TCP_PORT1=$(E2E_HOSTPORT_CONTROLVM_PROXY_SERVER) \
		  HOSTPORT_HOSTVM_TCP_PORT1=$(E2E_HOSTPORT_HOSTVM_PROXY_SERVER) \
		  KUBECONFIG_PATH=$(E2E_KUBECONFIG_PATH) \
		  HOSTPORT_HOSTVM_TCP_PORT2=$(E2E_PYROSCOPE_HOST_PORT)  \
		  VMPORT_HOSTVM_PYROSCOPE=$(E2E_PYROSCOPE_VM_PORT)  \
		  VAGRANT_IMAGE_K8S=$(E2E_VAGRANT_IMAGE_K8S) \
		  VAGRANT_IMAGE_UBUNTU=$(E2E_VAGRANT_IMAGE_UBUNTU) \
		  ./setupCluster.sh on
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
	if [ -n "$(E2E_PYROSCOPE_VM_PORT)" ] ; then \
  		echo "install proscope on vm $(E2E_VM_HOST)  " ; \
  		( cd vagrant ; ./ssh "$(E2E_VM_HOST)" "sudo docker stop $(PYROSCOPE_CONTAINER_NAME) &>/dev/null || true" ) ; \
  		( cd vagrant ; ./ssh "$(E2E_VM_HOST)" "sudo docker rm $(PYROSCOPE_CONTAINER_NAME) &>/dev/null || true" ) ; \
  		( cd vagrant ; ./ssh "$(E2E_VM_HOST)" "sudo docker run -d --name $(PYROSCOPE_CONTAINER_NAME) -p $(E2E_PYROSCOPE_VM_PORT):4040 $(PYROSCOPE_IMAGE_NAME) server" ) ; \
		echo "finish setuping pyroscope " ; \
		HOST_IP=$$( ip r get 8.8.8.8 | grep -oE "src [\.0-9]+" | grep -oE "[\.0-9]+" ) ; \
		echo "visit pyroscope: http://$${HOST_IP}:$(E2E_PYROSCOPE_HOST_PORT)" ; \
      fi

.PHONY: showInfo
showInfo:
	@HOST_IP=$$( ip r get 8.8.8.8 | grep -oE "src [\.0-9]+" | grep -oE "[\.0-9]+" ) ; \
		echo "visit pyroscope: http://$${HOST_IP}:$(E2E_PYROSCOPE_HOST_PORT)"
		echo "visit proxy server on $(E2E_VM_HOST): http://127.0.0.1:$(E2E_HOSTPORT_HOSTVM_PROXY_SERVER)"
		echo "visit proxy server on master node: http://127.0.0.1:$(E2E_HOSTPORT_CONTROLVM_PROXY_SERVER)"
	@echo "========================================================"
	@echo "   export KUBECONFIG=$(E2E_KUBECONFIG_PATH)             "
	@echo "   kubectl get pod -o wide -A                           "
	@echo "========================================================"

#==================

.PHONY: check_test_app_images_ready
check_test_app_images_ready:
	CHART_LIST=" $(ROOT_DIR)/tests/charts/backendserver $(ROOT_DIR)/tests/charts/proxyserver $(ROOT_DIR)/tests/charts/redirectserver " ;\
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
					EXIST=`sudo docker images | awk '{printf("%s:%s\n",$$1,$$2)}' | grep "$${IMAGE}" ` || true ; \
					if [ -z "$${EXIST}" ] ; then \
							CI_IMAGE=$${IMAGE##*/} ; \
							echo "try to find github CI image $${CI_IMAGE} " ; \
							EXIST=`sudo docker images | awk '{printf("%s:%s\n",$$1,$$2)}' | grep "$${CI_IMAGE}" ` || true ; \
							if [ -z "$${EXIST}" ] ; then \
								echo "error, failed to find image $${IMAGE}" ; \
								echo "error, failed to find image $${CI_IMAGE}" ; \
								echo "try to build image 'build_local_test_app_image' " ; \
								exit 1 ; \
							fi ; \
							sudo docker tag $${CI_IMAGE} $${IMAGE} ; \
					fi ;\
					echo "image exists: $${IMAGE}" ; \
				done ; \
				sudo docker images ; \
			fi \
		done

# this will auto tag github ci image : agent:xxx -> github.com/elf-io/balancing/agent:xxx
.PHONY: check_images_ready
check_images_ready:
	echo "check balancing image  " ; \
	IMAGE_LIST=` helm template test $(ROOT_DIR)/charts --set global.imageTagOverride=$(PROJECT_IMAGE_TAG)  | grep " image: " | tr -d '"'| awk '{print $$2}' ` ; \
	if [ -z "$${IMAGE_LIST}" ] ; then \
		echo "warning, failed to find image from chart " ; \
		exit 1 ;\
	else \
		echo "find image from chart : $${IMAGE_LIST} " ; \
		for IMAGE in $${IMAGE_LIST} ; do \
		  	echo "try to find image $${IMAGE} " ; \
			EXIST=`sudo docker images | awk '{printf("%s:%s\n",$$1,$$2)}' | grep "$${IMAGE}" ` || true ; \
			if [ -z "$${EXIST}" ] ; then \
					CI_IMAGE=$${IMAGE##*/} ; \
			  		echo "try to find github CI image $${CI_IMAGE} " ; \
			  		EXIST=`sudo docker images | awk '{printf("%s:%s\n",$$1,$$2)}' | grep "$${CI_IMAGE}" ` || true ; \
			  		if [ -z "$${EXIST}" ] ; then \
			  			echo "error, failed to find image $${IMAGE}" ; \
			  			echo "error, failed to find image $${CI_IMAGE}" ; \
			  			exit 1 ; \
			  		fi ; \
			  		sudo docker tag $${CI_IMAGE} $${IMAGE} ; \
			fi ;\
			echo "image exists: $${IMAGE}" ; \
		done ; \
		sudo docker images ; \
	fi


# install balancing on global cluster
.PHONY: deploy_project
deploy_project:
	echo "try to load local image tag $(PROJECT_IMAGE_TAG) " ; \
		IMAGE_LIST=` helm template test $(ROOT_DIR)/charts --set global.imageTagOverride=$(PROJECT_IMAGE_TAG)  | grep " image: " | tr -d '"'| awk '{print $$2}' ` ; \
		if [ -z "$${IMAGE_LIST}" ] ; then \
			echo "warning, failed to find image from chart " ; \
		else \
			echo "found image from chart : $${IMAGE_LIST} " ; \
			for IMAGE in $${IMAGE_LIST} ; do \
				EXIST=`sudo docker images | awk '{printf("%s:%s\n",$$1,$$2)}' | grep "$${IMAGE}" ` ; \
				if [ -z "$${EXIST}" ] ; then \
				  echo "docker pull $${IMAGE} to local" ; \
				  sudo docker pull $${IMAGE} ; \
				fi ;\
				echo "load local image $${IMAGE} " ; \
				( cd vagrant ; ./loadHostImages.sh $${IMAGE}  ) ; \
				if grep "agent" <<< "$${IMAGE}" &>/dev/null ; then \
						AGENT_IMAGE=$${IMAGE} ; \
				fi ; \
			done ; \
		fi  ; \
		echo "begin to deploy on kubernetes" ; \
		helm --kubeconfig=$(E2E_KUBECONFIG_PATH) uninstall -n $(E2E_INSTALL_NAMESPACE) project || true  ;\
		HELM_OPTION="" ; \
    	if [ -n "$(E2E_PYROSCOPE_VM_PORT)" ] ; then \
			echo "add env" ; \
			HELM_OPTION+=" --set balancingAgent.extraEnv[0].name=ENV_PYROSCOPE_PUSH_SERVER_ADDRESS  --set balancingAgent.extraEnv[0].value=http://192.168.0.2:$(E2E_PYROSCOPE_VM_PORT) " ; \
			HELM_OPTION+=" --set balancingController.extraEnv[0].name=ENV_PYROSCOPE_PUSH_SERVER_ADDRESS  --set balancingController.extraEnv[0].value=http://192.168.0.2:$(E2E_PYROSCOPE_VM_PORT) " ; \
		fi ; \
		HELM_OPTION+=" --set balancingAgent.debug.logLevel=debug --set balancingController.debug.logLevel=debug " ; \
		HELM_OPTION+=" --set balancingAgent.prometheus.enabled=true --set balancingController.prometheus.enabled=true  " ; \
		HELM_OPTION+=" --set balancingAgent.nodeEntryInterface=eth1 " ; \
		helm --kubeconfig=$(E2E_KUBECONFIG_PATH) install project $(ROOT_DIR)/charts \
				-n $(E2E_INSTALL_NAMESPACE) --create-namespace --wait --debug \
				--set global.imageTagOverride=$(PROJECT_IMAGE_TAG) \
				--set feature.apiServerHost=192.168.0.10 \
                --set feature.apiServerPort=6443 \
				$${HELM_OPTION} \
				|| {  ./scripts/debugCluster.sh $(E2E_KUBECONFIG_PATH) "detail"  $(E2E_INSTALL_NAMESPACE) ; exit 1 ; } ; \
		echo "begin to deploy agent on host node " ; \
		( cd vagrant ; ./cpToVM "$(E2E_VM_HOST)" ./config_old /tmp/config  ) ; \
		( cd vagrant ; ./ssh "$(E2E_VM_HOST)"  sudo docker stop balancing || true ; ) ; \
		( cd vagrant ; ./ssh "$(E2E_VM_HOST)" 	sudo docker rm balancing || true ; ) ; \
		( cd vagrant ; ./ssh "$(E2E_VM_HOST)"  \
			sudo docker run -d --net=host --privileged --name balancing \
						 -e "KUBECONFIG=/config"  -v /tmp/config:/config  \
                          -v /sys/fs:/sys/fs:rw  -v /proc:/host/proc \
                          "$${AGENT_IMAGE}"   ; \
		) ; \
		exit 0
		make showInfo

.PHONY: install_example_app
install_example_app:
	@echo "---------- install applicatio on kubernetes"
	echo "get images ... " ; \
		CHART_LIST=" $(ROOT_DIR)/tests/charts/backendserver $(ROOT_DIR)/tests/charts/proxyserver $(ROOT_DIR)/tests/charts/redirectserver " ;\
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
					EXIST=`sudo docker images | awk '{printf("%s:%s\n",$$1,$$2)}' | grep "$${IMAGE}" ` || true ; \
					if [ -z "$${EXIST}" ] ; then \
						  echo "docker pull $${IMAGE} to local" ; \
						  sudo docker pull $${IMAGE} ; \
					fi ;\
					echo "load local image $${IMAGE} " ; \
				    ( cd vagrant ; ./loadHostImages.sh $${IMAGE}  ) ; \
				done ; \
			fi \
		done
	echo "deploy proxy server " ; \
			helm  --kubeconfig=$(E2E_KUBECONFIG_PATH) uninstall  proxyserver &>/dev/null || true ; \
			helm  --kubeconfig=$(E2E_KUBECONFIG_PATH) install  proxyserver ./charts/proxyserver \
					--wait --debug --set image.tag=$(TEST_APP_IMAGE_TAG) \
			|| { ./scripts/debugCluster.sh $(E2E_KUBECONFIG_PATH) "detail"  $(E2E_INSTALL_NAMESPACE) ; exit 1 ; } ;\
			echo " deploy proxy on host node" ; \
			IMAGE=` helm template test ./charts/proxyserver --set image.tag=$(TEST_APP_IMAGE_TAG)  | grep " image: " | tr -d '"'| awk '{print $$2}' ` ; \
			( cd vagrant ; ./ssh "$(E2E_VM_HOST)"  sudo docker stop server || true ; ) ; \
			( cd vagrant ; ./ssh "$(E2E_VM_HOST)" sudo docker rm server || true ; ) ; \
			( cd vagrant ; ./ssh "$(E2E_VM_HOST)" \
				sudo docker run -d -p 80:80 --name server "$${IMAGE}" ; \
			)
	echo "deploy backend server " ; \
			helm  --kubeconfig=$(E2E_KUBECONFIG_PATH) uninstall  backendserver &>/dev/null || true ; \
			helm  --kubeconfig=$(E2E_KUBECONFIG_PATH) install  backendserver ./charts/backendserver \
			--wait --debug --set image.tag=$(TEST_APP_IMAGE_TAG) \
			|| {   ./scripts/debugCluster.sh $(E2E_KUBECONFIG_PATH) "detail"  $(E2E_INSTALL_NAMESPACE) ; exit 1 ; }
	echo "deploy redirect server " ; \
			helm  --kubeconfig=$(E2E_KUBECONFIG_PATH) uninstall  redirectserver &>/dev/null || true ; \
			helm  --kubeconfig=$(E2E_KUBECONFIG_PATH) install  redirectserver ./charts/redirectserver \
			--wait --debug --set image.tag=$(TEST_APP_IMAGE_TAG) \
			|| { ./scripts/debugCluster.sh $(E2E_KUBECONFIG_PATH) "detail"  $(E2E_INSTALL_NAMESPACE) ; exit 1 ; }
	@ KUBECONFIG=$(E2E_KUBECONFIG_PATH)  kubectl get pod -o wide -A
	@echo "========================================================"
	@echo "   export KUBECONFIG=$(E2E_KUBECONFIG_PATH)             "
	@echo "   kubectl get pod -o wide -A                           "
	@echo "   visit the proxy server on kubernetes : "
	@echo "                curl http://127.0.0.1:$(E2E_HOSTPORT_HOSTVM_PROXY_SERVER) | jq .                          "
	@echo "   visit the proxy server on host : "
	@echo "                curl http://127.0.0.1:$(E2E_HOSTPORT_CONTROLVM_PROXY_SERVER) | jq .                          "
	@echo "========================================================"


.PHONY: test_connectivity
test_connectivity:
	export K8S_PROXY_SERVER_MAPPING_PORT=$(E2E_HOSTPORT_CONTROLVM_PROXY_SERVER) ; \
 		export HOST_PROXY_SERVER_MAPPING_PORT=$(E2E_HOSTPORT_HOSTVM_PROXY_SERVER) ;\
 		export KUBECONFIG=$(E2E_KUBECONFIG_PATH) ; \
 		./scripts/testAppConnectivity.sh basic || { echo "failed to test the connectivity" ; exit 1 ; } ;  \
	    ./scripts/testAppConnectivity.sh service || { echo "failed to test the connectivity" ; exit 1 ; } ; \
	    ./scripts/testAppConnectivity.sh balancing || { echo "failed to test the connectivity" ; exit 1 ; } ; \
	    ./scripts/testAppConnectivity.sh redirect || { echo "failed to test the connectivity" ; exit 1 ; }


.PHONY: clean
clean:
	-@ cd ./vagrant ; ./setupCluster.sh off || true
	-@ rm -rf $(E2E_RUNTIME_DIR) || true
