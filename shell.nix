let
  sources = import ./nix/sources.nix;
  pkgs = import sources.nixpkgs { };
in
pkgs.mkShell {
  name = "lvm-shell";
  buildInputs = with pkgs; [
    chart-testing
    git
    go_1_24
    golangci-lint
    golint
    kubectl
    kubernetes-helm
    gnumake
    helm-docs
    minikube
    semver-tool
    yq-go
    which
    curl
    crane
    cacert
    util-linux
    jq
    lvm2_dmeventd
    nixos-shell
    niv
  ] ++ pkgs.lib.optional (builtins.getEnv "IN_NIX_SHELL" == "pure") [ docker-client ];

  PRE_COMMIT_ALLOW_NO_CONFIG = 1;

  shellHook = ''
    unset GOARCH
    unset GOOS
    unset GOROOT
    
    # Temp directories should not be in project directory to avoid issues if it's mounted remotely
    if [[ -z "$HOME" ]]; then
      export OPENEBS_CACHE="/tmp/.cache/openebs/lvm-localpv"
    else
      export OPENEBS_CACHE="$HOME/.cache/openebs/lvm-localpv"
    fi

    export CGO_ENABLED=0
    export GOCACHE="$OPENEBS_CACHE/go/cache"
    export GOENV=off
    export GOMODCACHE="$OPENEBS_CACHE/go/modcache"
    export GOPATH="$OPENEBS_CACHE/go"
    export GOPROXY=direct
    export GOTELEMETRY="off"
    export GOTMPDIR="$OPENEBS_CACHE/go/.tmp"
    export GOTOOLCHAIN=local
    export PATH="$GOPATH/bin:$PATH"
    export TMPDIR="$OPENEBS_CACHE/tmp"

    mkdir -p "$GOCACHE"
    mkdir -p "$GOMODCACHE"
    mkdir -p "$GOTMPDIR"
    mkdir -p "$TMPDIR"

    if [ "$IN_NIX_SHELL" = "pure" ]; then
      # working sudo within a pure nix-shell
      for sudo in /run/wrappers/bin/sudo /usr/bin/sudo /usr/local/bin/sudo /sbin/sudo /bin/sudo; do
        mkdir -p $OPENEBS_CACHE/bins
        ln -sf $sudo $OPENEBS_CACHE/bins/sudo
        export PATH=$OPENEBS_CACHE/bins:$PATH
        break
      done
    else
      rm $OPENEBS_CACHE/bins/sudo 2>/dev/null || :
      rmdir $OPENEBS_CACHE/bins 2>/dev/null || :
    fi

    make bootstrap
  '';
}
