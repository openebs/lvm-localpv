# Vagrantfile to spin up a development VM
# 
# Prerequisites:
# - Install VirtualBox: https://www.virtualbox.org/
# - Start VM: `vagrant up`
# - Install VSCode Remote-SSH extension: https://marketplace.visualstudio.com/items?itemName=ms-vscode-remote.remote-ssh
# - Print SSH config: `vagrant ssh-config`
# - Manually add SSH config to your `~/.ssh/config` and change the name from `default` to `lvm-localpv`
# - Open the project remotely with: `code --folder-uri 'vscode-remote://ssh-remote+lvm-localpv/workspace'

require "json"

GOLANGCI_LINT_VERSION = File.read("Makefile").match(%r{^\s*GOLANGCI_LINT_VERSION\s*=\s*(\d+(?:\.\d+)*)})[1]
GO_VERSION = File.read(".github/workflows/build_and_push.yml").match(%r{^\s*(?:go-version:\s+)(\d+(?:\.\d+)*)})[1]

LOCAL_ZSH = <<~TEXT
  # Setup zsh-snap
  if [[ ! -r ~/.config/znap/znap/znap.zsh ]]; then
      git clone https://github.com/marlonrichert/zsh-snap.git ~/.config/znap/znap
  fi

  source ~/.config/znap/znap/znap.zsh

  # Configure Starship
  znap eval starship "starship init zsh"

  # Configure Docker
  znap fpath _docker "docker completion zsh"

  # Configure kubectl
  znap fpath _kubectl "kubectl completion zsh"

  # Configure kubeswitch
  znap eval switcher "switcher init zsh"
  znap fpath _switcher "switcher completion zsh"
  alias s=switcher

  # Configure Nix
  if [ -r /home/vagrant/.nix-profile/etc/profile.d/nix.sh ]; then
    znap eval nix "cat /home/vagrant/.nix-profile/etc/profile.d/nix.sh"
  fi

  # Configure environment
  export PATH=${PATH}:/usr/local/go/bin:~/go/bin
  export EDITOR=emacs
  export CGO_ENABLED=0
TEXT

LVM_CONF = <<~TEXT
  activation {
      thin_pool_autoextend_threshold = 50
      thin_pool_autoextend_percent = 20
  }
TEXT

Vagrant.configure("2") do |config|
  config.vm.box = "bento/debian-13"
  config.vm.hostname = "lvm-localpv"
  config.vm.box_check_update = false
  config.vm.disk :disk, size: "256GB", primary: true
  
  config.vm.provider "virtualbox" do |provider|
    provider.name = "lvm-localpv"
    provider.cpus = 4
    provider.memory = 16 * 1024
    provider.customize ["modifyvm", :id, "--natdnshostresolver1", "on"]
    provider.customize ["modifyvm", :id, "--natdnsproxy1", "on"]
  end

  config.vm.synced_folder ".", "/workspace", type: "virtualbox"

  config.vm.provision "shell", name: "System Setup", inline: <<~SHELL
    set -e

    # Version configuration
    CRI_DOCKERD_VERSION="0.3.20"
    DIVE_VERSION="0.13.1"
    K9S_VERSION="0.50.13"
    KUBERNETES_VERSION="1.34.0"
    MINIKUBE_VERSION="1.37.0"
    ACT_VERSION="0.2.82"
    KUBERNETES_MINOR_VERSION=$(echo "${KUBERNETES_VERSION}" | cut -d. -f1-2)

    # APT packages to install
    APT_PACKAGES="conntrack containerd.io cri-tools docker-buildx-plugin docker-ce docker-ce-cli docker-compose-plugin emacs-nox git helm jq kubectl kubernetes-cni make starship thin-provisioning-tools zsh"

    # Function to download files with caching
    download_file() {
      local url="$1"
      local output_file="$2"
      
      if [ ! -f "${output_file}" ]; then
        curl --fail --silent --show-error --location "${url}" --output "${output_file}"
      fi
    }

    # Function to add line to file if not already present
    add_line_to_file() {
      local file="$1"
      local line="$2"
      
      if ! grep -q "${line}" "${file}"; then
        printf "${line}" >>"${file}"
      fi
    }

    # Function to write content to file
    write_to_file() {
      local file="$1"
      local content="$2"
      
      printf "%s\n" "${content}" >"${file}"
    }

    # Configure systemd-timesyncd
    write_to_file "/etc/systemd/timesyncd.conf" "[Time]\nNTP = pool.ntp.org"
    systemctl restart systemd-timesyncd

    # Create download directory
    mkdir --parents /var/cache/downloads

    # Detect environment
    ARCH=$(dpkg --print-architecture)
    CODENAME=$(lsb_release --codename --short)

    # Setup Docker repository
    download_file "https://download.docker.com/linux/debian/gpg" "/var/cache/downloads/docker.asc"
    
    cp /var/cache/downloads/docker.asc /usr/share/keyrings/docker.asc
    write_to_file "/etc/apt/sources.list.d/docker.list" "deb [arch=${ARCH} signed-by=/usr/share/keyrings/docker.asc] https://download.docker.com/linux/debian ${CODENAME} stable"

    # Setup Kubernetes repository
    download_file "https://pkgs.k8s.io/core:/stable:/v${KUBERNETES_MINOR_VERSION}/deb/Release.key" "/var/cache/downloads/kubernetes-${KUBERNETES_MINOR_VERSION}.asc"
    
    cp /var/cache/downloads/kubernetes-${KUBERNETES_MINOR_VERSION}.asc /usr/share/keyrings/kubernetes.asc
    write_to_file "/etc/apt/sources.list.d/kubernetes.list" "deb [signed-by=/usr/share/keyrings/kubernetes.asc] https://pkgs.k8s.io/core:/stable:/v${KUBERNETES_MINOR_VERSION}/deb/ /"

    # Setup Helm repository
    download_file "https://packages.buildkite.com/helm-linux/helm-debian/gpgkey" "/var/cache/downloads/helm.asc"
    
    cp /var/cache/downloads/helm.asc /usr/share/keyrings/helm.asc
    write_to_file "/etc/apt/sources.list.d/helm.list" "deb [arch=${ARCH} signed-by=/usr/share/keyrings/helm.asc] https://packages.buildkite.com/helm-linux/helm-debian/any/ any main"

    # Install required software. Ok, mostly required…
    apt-get update --quiet --quiet
    apt-get install --yes --quiet --quiet --no-install-recommends --option=Dpkg::Options::=--force-confdef --option=Dpkg::Options::=--force-confold ${APT_PACKAGES}

    # Configure sudo
    write_to_file "/etc/sudoers.d/keep_vars" 'Defaults env_keep += "EDITOR SSH_AGENT_PID SSH_AUTH_SOCK"'
    chmod 440 /etc/sudoers.d/keep_vars

    # Configure Docker
    usermod --append --groups docker vagrant
    
    # Install Go
    download_file "https://go.dev/dl/go#{GO_VERSION}.linux-${ARCH}.tar.gz" "/var/cache/downloads/go-#{GO_VERSION}.tar.gz"

    tar --directory /usr/local --extract --gzip --file /var/cache/downloads/go-#{GO_VERSION}.tar.gz --overwrite

    # Install golangci-lint
    download_file "https://github.com/golangci/golangci-lint/releases/download/v#{GOLANGCI_LINT_VERSION}/golangci-lint-#{GOLANGCI_LINT_VERSION}-linux-${ARCH}.tar.gz" "/var/cache/downloads/golangci-lint-#{GOLANGCI_LINT_VERSION}.tar.gz"

    tar --extract --gzip --file /var/cache/downloads/golangci-lint-#{GOLANGCI_LINT_VERSION}.tar.gz --strip-components 1 --directory /usr/local/bin --wildcards "*/golangci-lint" --overwrite

    download_file "https://raw.githubusercontent.com/Mirantis/cri-dockerd/v${CRI_DOCKERD_VERSION}/packaging/systemd/cri-docker.service" "/var/cache/downloads/cri-docker-${CRI_DOCKERD_VERSION}.service"
    cp /var/cache/downloads/cri-docker-${CRI_DOCKERD_VERSION}.service /etc/systemd/system/cri-docker.service
    sed --in-place --expression='s,/usr/bin/cri-dockerd,/usr/local/bin/cri-dockerd,' /etc/systemd/system/cri-docker.service

    download_file "https://raw.githubusercontent.com/Mirantis/cri-dockerd/v${CRI_DOCKERD_VERSION}/packaging/systemd/cri-docker.socket" "/var/cache/downloads/cri-docker-${CRI_DOCKERD_VERSION}.socket"
    cp /var/cache/downloads/cri-docker-${CRI_DOCKERD_VERSION}.socket /etc/systemd/system/cri-docker.socket

    # Install kubeswitch
    download_file "https://github.com/danielfoehrKn/kubeswitch/releases/download/0.9.3/switcher_linux_${ARCH}" "/var/cache/downloads/kubeswitch-0.9.3"

    cp /var/cache/downloads/kubeswitch-0.9.3 /usr/local/bin/switcher
    chmod a+x /usr/local/bin/switcher
    
    # Install cri-dockerd
    if systemctl is-active --quiet cri-docker 2>/dev/null; then
      systemctl stop cri-docker
    fi

    download_file "https://github.com/Mirantis/cri-dockerd/releases/download/v${CRI_DOCKERD_VERSION}/cri-dockerd-${CRI_DOCKERD_VERSION}.${ARCH}.tgz" "/var/cache/downloads/cri-dockerd-${CRI_DOCKERD_VERSION}.tar.gz"
    tar --extract --gzip --file /var/cache/downloads/cri-dockerd-${CRI_DOCKERD_VERSION}.tar.gz --strip-components 1 --directory /usr/local/bin --overwrite

    systemctl daemon-reload
    systemctl enable cri-docker.service
    systemctl enable --now cri-docker.socket

    # Install k9s
    download_file "https://github.com/derailed/k9s/releases/download/v${K9S_VERSION}/k9s_linux_${ARCH}.deb" "/var/cache/downloads/k9s-${K9S_VERSION}.deb"
    dpkg --install /var/cache/downloads/k9s-${K9S_VERSION}.deb

    # Install dive
    download_file "https://github.com/wagoodman/dive/releases/download/v${DIVE_VERSION}/dive_${DIVE_VERSION}_linux_${ARCH}.deb" "/var/cache/downloads/dive-${DIVE_VERSION}.deb"
    dpkg --install /var/cache/downloads/dive-${DIVE_VERSION}.deb

    # Install and configure minikube
    download_file "https://github.com/kubernetes/minikube/releases/download/v${MINIKUBE_VERSION}/minikube_${MINIKUBE_VERSION}-0_${ARCH}.deb" "/var/cache/downloads/minikube-${MINIKUBE_VERSION}.deb"

    dpkg --install /var/cache/downloads/minikube-${MINIKUBE_VERSION}.deb
    sysctl fs.protected_regular=0
    write_to_file "/etc/sysctl.d/minikube.conf" "fs.protected_regular=0"

    minikube config set driver none
    minikube config set kubernetes-version "${KUBERNETES_VERSION}"
    minikube start --download-only 

    # Install act (Github local action runner)
    ACT_ARCH=$(case "$ARCH" in
      amd64) echo "x86_64" ;;
      arm64) echo "arm64" ;;
      *) echo "unsupported"; exit 1 ;;
    esac)

    download_file "https://github.com/nektos/act/releases/download/v${ACT_VERSION}/act_Linux_${ACT_ARCH}.tar.gz" "/var/cache/downloads/act-${ACT_VERSION}.tar.gz"
    tar --extract --gzip --file /var/cache/downloads/act-${ACT_VERSION}.tar.gz --directory /usr/local/bin act --overwrite

    # Configure LVM2
    write_to_file "/etc/lvm/lvm.conf" '#{LVM_CONF}'
    systemctl daemon-reload
    systemctl enable --now lvm2-monitor.service

    # Configure zsh
    GLOBAL_ZSHRC="/etc/zsh/zshrc"
    ZSHRC="/etc/zsh/zshrc.d/local.zsh"

    mkdir --parents /etc/zsh/zshrc.d

    add_line_to_file "${GLOBAL_ZSHRC}" "[ -r \"${ZSHRC}\" ] && source \"${ZSHRC}\""

    write_to_file "/etc/zsh/zshrc.d/local.zsh" '#{LOCAL_ZSH}'

    touch "${HOME}/.zshrc"

    if [ ! -r "${HOME}/.config/znap/znap" ]; then
      git clone "https://github.com/marlonrichert/zsh-snap.git" "${HOME}/.config/znap/znap"
    fi

    for user in root vagrant; do
      home_dir=$(getent passwd ${user} | cut -d: -f6)

      chsh --shell /bin/zsh ${user}
      install --owner=${user} --group=${user} --mode=0644 /dev/null "${home_dir}/.zshrc"

      if [ ! -r "${home_dir}/.config/znap/znap" ]; then
        sudo --user=${user} -- git clone "https://github.com/marlonrichert/zsh-snap.git" "${home_dir}/.config/znap/znap"
      fi
    done
  SHELL

  config.vm.provision "shell", name: "User Setup", privileged: false, inline: <<~SHELL
    # Install Nix
    if ! command -v nix >/dev/null 2>&1; then
      curl -L https://nixos.org/nix/install | sh
    fi
  SHELL

  config.vm.post_up_message = "Use `code --folder-uri 'vscode-remote://ssh-remote+#{config.vm.hostname}/workspace'` to start vscode with Remote-SSH"
end
