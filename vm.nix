{ ... }:
let
  sources = import ./nix/sources.nix;
  pkgs = import sources.nixpkgs { };
in
{
  nix.nixPath = [
    "nixpkgs=${pkgs.path}"
  ];
  nixos-shell.mounts = {
    mountHome = false;
    mountNixProfile = false;
    cache = "none"; # default is "loose"

    extraMounts = {
      "/lvm" = {
        target = ./.;
        cache = "none";
      };
    };
  };

  virtualisation = {
    cores = 4;
    memorySize = 2048;
    # Uncomment to be able to ssh into the vm, example:
    # ssh -p 2222 -o StrictHostKeychecking=no root@localhost
    # forwardPorts = [
    #   { from = "host"; host.port = 2222; guest.port = 22; }
    # ];
    diskSize = 20 * 1024;
    docker = {
      enable = true;
    };
  };
  documentation.enable = false;

  networking.firewall = {
    allowedTCPPorts = [
      6443 # k3s: required so that pods can reach the API server (running on port 6443 by default)
    ];
  };

  services = {
    openssh.enable = true;
    k3s = {
      enable = true;
      role = "server";
      extraFlags = toString [
        "--disable=traefik"
      ];
    };
    lvm = {
      dmeventd.enable = true;
    };
  };

  programs.git = {
    enable = true;
    config = {
      safe = { directory = "/lvm"; };
    };
  };

  environment = {
    variables = {
      CGO_ENABLED = "0";
      CI_K3S = "true";
      EDITOR = "vim";
      GOCACHE = "/var/cache/go/cache";
      GOENV = "off";
      GOMODCACHE="/var/cache/go/modcache";
      GOPATH = "/usr/share/go";
      GOPROXY = "direct";
      GOTMPDIR="/tmp";
      GOTELEMETRY="off";
      GOTOOLCHAIN = "local";
      KUBECONFIG = "/etc/rancher/k3s/k3s.yaml";
    };

    shellAliases = {
      k = "kubectl";
      ke = "kubectl -n openebs";
    };

    etc."lvm/lvm.conf".text = ''
      global {
        # system_id_source = "machineid"
      }
      activation {
        thin_pool_autoextend_threshold = 50
        thin_pool_autoextend_percent   = 20
      }
    '';

    shellInit = ''
      export PATH=$GOPATH/bin:$PATH
      cd /lvm
    '';

    systemPackages = with pkgs; [ vim docker-client k9s e2fsprogs go_1_24 ] ++ [ thin-provisioning-tools lvm2_dmeventd ];
  };
}
