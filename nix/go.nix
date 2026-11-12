{ pkgs ? import (import ./sources.nix).nixpkgs { } }:
let
  inherit (pkgs) lib;

  # Go major.minor comes from go.mod: toolchain directive first, then the go directive.
  goModLines = lib.splitString "\n" (builtins.readFile ../go.mod);
  goModMatch = re: lib.findFirst (m: m != null) null (map (builtins.match re) goModLines);
  goVersion =
    let toolchain = goModMatch "toolchain go([0-9]+)\\.([0-9]+).*";
    in if toolchain != null then toolchain else goModMatch "go ([0-9]+)\\.([0-9]+).*";
  goAttr = "go_${lib.concatStringsSep "_" goVersion}";
  go = pkgs.${goAttr} or (throw "nixpkgs has no ${goAttr} (from go.mod), update nix/sources.json");
in
pkgs.mkShell {
  name = "openebs-go";
  buildInputs = [
    go
    pkgs.golangci-lint
    pkgs.golint
  ];

  passthru = { inherit go; };

  shellHook = import ./common.nix + ''
    unset GOARCH
    unset GOOS
    unset GOROOT

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

    mkdir -p "$GOCACHE"
    mkdir -p "$GOMODCACHE"
    mkdir -p "$GOTMPDIR"
  '';
}
