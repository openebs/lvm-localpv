let
  sources = import ../nix/sources.nix;
  pkgs = import sources.nixpkgs { };
in
pkgs.mkShellNoCC {
  name = "lvm-shell";
  buildInputs = with pkgs; [
    chart-testing
    kubernetes-helm
    helm-docs
    semver-tool
    yq-go
  ];

  PRE_COMMIT_ALLOW_NO_CONFIG = 1;

  shellHook = ''
    # Temp directories should not be in project directory to avoid issues if it's mounted remotely
    if [[ -z "$HOME" ]]; then
      export OPENEBS_CACHE="/tmp/.cache/openebs/lvm-localpv"
    else
      export OPENEBS_CACHE="$HOME/.cache/openebs/lvm-localpv"
    fi

    export TMPDIR="$OPENEBS_CACHE/tmp"

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
  '';
}
