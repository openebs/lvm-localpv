{ pkgs ? import (import ./sources.nix).nixpkgs { } }:
pkgs.mkShellNoCC {
  name = "lvm-chart-shell";
  buildInputs = with pkgs; [
    chart-testing
    kubernetes-helm
    helm-docs
    semver-tool
    yq-go
  ];

  shellHook = import ./common.nix;
}
