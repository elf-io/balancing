#！/bin/bash
## SPDX-License-Identifier: Apache-2.0
## Copyright Authors of Spider

OS=$(uname | tr 'A-Z' 'a-z')

DOWNLOAD_OPT=""
if [ -n "$http_proxy" ]; then
  DOWNLOAD_OPT=" -x $http_proxy "
  export http_proxy="$http_proxy"
fi

if ! kubectl help &>/dev/null  ; then
    echo "error, miss 'kubectl', try to install it "
    LATEST_VERSION=$(curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt)
    echo "downland kubectl ${LATEST_VERSION}"
    curl ${DOWNLOAD_OPT} -Lo /usr/local/bin/kubectl https://storage.googleapis.com/kubernetes-release/release/${LATEST_VERSION}/bin/$OS/amd64/kubectl
    chmod +x /usr/local/bin/kubectl
    ! kubectl -h  &>/dev/null && echo "error, failed to install kubectl" && exit 1
else
    echo "pass   'kubectl' installed:  $(kubectl version --client=true | grep -E -o "Client.*GitVersion:\"[^[:space:]]+\"" | awk -F, '{print $NF}') "
fi

# Install Helm
if ! helm > /dev/null 2>&1 ; then
    echo "error, miss 'helm', try to install it "
    LATEST_VERSION=` curl -s https://api.github.com/repos/helm/helm/releases/latest |  grep -Po '"tag_name": "\K.*?(?=")' `
    if [ -z "$LATEST_VERSION" ] ; then
        echo "error, failed to get latest version for helm"
        exit 1
    fi
    curl ${DOWNLOAD_OPT} -Lo /tmp/helm.tar.gz "https://get.helm.sh/helm-${LATEST_VERSION}-$OS-amd64.tar.gz"
    tar -xzvf /tmp/helm.tar.gz && mv $OS-amd64/helm  /usr/local/bin
    chmod +x /usr/local/bin/helm
    rm /tmp/helm.tar.gz
    rm $OS-amd64/LICENSE
    rm $OS-amd64/README.md
    ! helm version &>/dev/null && echo "error, failed to install helm" && exit 1
else
    echo "pass   'helm' installed:  $( helm version | grep -E -o "Version:\"v[^[:space:]]+\"" ) "
fi


if ! VBoxManage -v &>/dev/null ; then
    echo "install virtual box"
    #https://www.virtualbox.org/wiki/Linux_Downloads
    wget https://download.virtualbox.org/virtualbox/7.0.20/virtualbox-7.0_7.0.20-163906~Ubuntu~jammy_amd64.deb
    apt-get update  -y
    apt install ./virtualbox-7.0_7.0.20-163906~Ubuntu~jammy_amd64.deb
    VBoxManage -v
else
    echo "virtual box is ready"
fi

if ! vagrant version &>/dev/null ; then
    # https://developer.hashicorp.com/vagrant/downloads
    wget -O- https://apt.releases.hashicorp.com/gpg | sudo gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg
    echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/hashicorp.list
    sudo apt update && sudo apt install -y vagrant
    vagrant version
else
    echo "vagrant is ready"
fi


if ! which sshpass &>/dev/null ; then
    apt-get install -y sshpass
fi
if ! which jq &>/dev/null ; then
    apt-get install -y jq
fi

exit 0
