{ pkgs ? import (import ./sources.nix).nixpkgs { } }:
let
  # Upstream's static release binaries, with the sha256 GitHub lists for each asset.
  assets = {
    x86_64-linux = { name = "linux-amd64"; sha256 = "f90880b27b586529acab8b950b65bd2793557e4a679672cfc2376b9aeaeee138"; };
    aarch64-linux = { name = "linux-arm64"; sha256 = "f6af4e62e46e34e121128a6d98705e7cf7575dd10083eb828dc6ad518754e0b3"; };
    x86_64-darwin = { name = "darwin-amd64"; sha256 = "4a0c23121b60d89e78b2c1d667a78d4b348f8a83a91cde09441139f85b6e1329"; };
    aarch64-darwin = { name = "darwin-arm64"; sha256 = "c91ef2a8f9fa346b42b626145646abbb7f8a5aa8f8b39e49c083bd630c875c7d"; };
  };
  asset = assets.${pkgs.stdenv.hostPlatform.system} or (throw "no preflight release binary for ${pkgs.stdenv.hostPlatform.system}");
  preflight = pkgs.stdenvNoCC.mkDerivation rec {
    pname = "preflight";
    version = "1.23.0";
    src = pkgs.fetchurl {
      url = "https://github.com/redhat-openshift-ecosystem/openshift-preflight/releases/download/${version}/preflight-${asset.name}";
      inherit (asset) sha256;
    };
    dontUnpack = true;
    installPhase = ''
      install -Dm755 $src $out/bin/preflight
    '';
  };
in
pkgs.mkShellNoCC {
  name = "lvm-preflight-shell";
  buildInputs = with pkgs; [
    preflight
    skopeo
    jq
  ];

  passthru = { inherit preflight; };

  shellHook = import ./common.nix;
}
