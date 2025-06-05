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
    echo "pass   'kubectl' installed:  $( kubectl version --client=true ) "
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


if ! sudo VBoxManage -v &>/dev/null ; then
    echo "install virtual box"
    #https://www.virtualbox.org/wiki/Linux_Downloads
    
    # Detect Ubuntu version and use appropriate package
    UBUNTU_VERSION=$(lsb_release -cs)
    echo "Detected Ubuntu version: $UBUNTU_VERSION"
    cat /etc/os-release
    
    if [ "$UBUNTU_VERSION" = "noble" ]; then
        # For Ubuntu 24.04 (Noble), use the official repository method
        echo "Using VirtualBox repository for Ubuntu $UBUNTU_VERSION"
        # Add Oracle VirtualBox public key
        wget -O- https://www.virtualbox.org/download/oracle_vbox_2016.asc | sudo gpg --dearmor --yes --output /usr/share/keyrings/oracle-virtualbox-2016.gpg
        
        # Add the VirtualBox repository
        echo "deb [arch=amd64 signed-by=/usr/share/keyrings/oracle-virtualbox-2016.gpg] https://download.virtualbox.org/virtualbox/debian $UBUNTU_VERSION contrib" | sudo tee /etc/apt/sources.list.d/virtualbox.list
        
        # Update and install VirtualBox
        sudo apt-get update -y
        sudo apt-get install -y virtualbox-7.0
    else
        # For Ubuntu 22.04 (Jammy) and others, use the direct package download
        echo "Using direct package download for Ubuntu $UBUNTU_VERSION"
        [ -f ./virtualbox-7.0_7.0.20-163906~Ubuntu~jammy_amd64.deb ] || wget https://download.virtualbox.org/virtualbox/7.0.20/virtualbox-7.0_7.0.20-163906~Ubuntu~jammy_amd64.deb
        sudo apt-get update -y
        sudo apt install -y ./virtualbox-7.0_7.0.20-163906~Ubuntu~jammy_amd64.deb
    fi
    
    # Install dependencies if needed
    sudo apt-get install -f -y
    
    sudo VBoxManage -v
else
    echo "pass virtual box is ready: $(vboxmanage --version) "
fi

if ! sudo vagrant version &>/dev/null ; then
    # https://developer.hashicorp.com/vagrant/downloads
    wget -O- https://apt.releases.hashicorp.com/gpg | sudo gpg --dearmor -o /usr/share/keyrings/hashicorp-archive-keyring.gpg
    echo "deb [signed-by=/usr/share/keyrings/hashicorp-archive-keyring.gpg] https://apt.releases.hashicorp.com $(lsb_release -cs) main" | sudo tee /etc/apt/sources.list.d/hashicorp.list
    sudo apt update && sudo apt install -y vagrant
    sudo vagrant version
    sudo cat /etc/vbox/networks.conf
    sudo echo '* 0.0.0.0/0 ::/0' > /etc/vbox/networks.conf
else
    echo "pass vagrant is ready: $(vagrant version) "
fi


if ! which sshpass &>/dev/null ; then
    sudo apt-get install -y sshpass
fi
if ! which jq &>/dev/null ; then
    sudo apt-get install -y jq
fi

exit 0
