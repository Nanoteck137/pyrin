{
  description = "Devshell for pyrin";

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

        pyrin = pkgs.buildGoModule {
          pname = "pyrin";
          version = self.shortRev or "dirty";
          src = ./.;

          vendorHash = "sha256-BkZx48XDdtMXEljRI01sJm/Kqov4PZW8TH+F11JpfvQ=";
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
