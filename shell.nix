{ bootstrap ? false }:
let
  sources = import ./nix/sources.nix;
  pkgs = import sources.nixpkgs { };
  go = import ./nix/go.nix { inherit pkgs; };
  chart = import ./nix/chart.nix { inherit pkgs; };
in
pkgs.mkShell {
  name = "lvm-shell";
  inputsFrom = [ go ];
  buildInputs = chart.buildInputs ++ (with pkgs; [
    git
    kubectl
    gnumake
    minikube
    which
    curl
    crane
    cacert
    util-linux
    jq
    lvm2_dmeventd
    nixos-shell
    niv
  ]) ++ pkgs.lib.optional (builtins.getEnv "IN_NIX_SHELL" == "pure") pkgs.docker-client;

  shellHook = ''
    if [ "${toString bootstrap}" = "1" ]; then make bootstrap; fi
  '';
}
