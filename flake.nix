{
  description = "Pyrin Code Generator";

  inputs = {
    nixpkgs.url      = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url  = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        overlays = [];
        pkgs = import nixpkgs {
          inherit system overlays;
        };

        version = pkgs.lib.strings.fileContents "${self}/version";
        fullVersion = ''${version}-${self.dirtyShortRev or self.shortRev or "dirty"}'';

        pyrin = pkgs.buildGoModule {
          pname = "pyrin";
          version = fullVersion;
          src = ./.;

          vendorHash = "sha256-jKYbQ54+bmLHej5IYg2YkreQa9xMVTLPTLRJY91v97M=";
        };
      in
      {
        packages.default = pyrin;
        packages.pyrin = pyrin;

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
          ];
        };
      }
    );
}
