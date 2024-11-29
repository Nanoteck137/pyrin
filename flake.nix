{
  description = "Pyrin Code Generator";

  inputs = {
    nixpkgs.url      = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url  = "github:numtide/flake-utils";

    devtools.url     = "github:nanoteck137/devtools";
    devtools.inputs.nixpkgs.follows = "nixpkgs";
  };

  outputs = { self, nixpkgs, flake-utils, devtools, ... }:
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
          subPackages = ["cmd/pyrin"];

          ldflags = [
            "-X github.com/nanoteck137/pyrin/cmd/pyrin/cli.Version=${version}"
            "-X github.com/nanoteck137/pyrin/cmd/pyrin/cli.Commit=${self.dirtyRev or self.rev or "no-commit"}"
          ];

          vendorHash = "sha256-TqgOO3v2n/Ot+HoXuXvhaa+3NuFQ/eAbx19MVillby0=";
        };

        tools = devtools.packages.${system};
      in
      {
        packages.default = pyrin;
        packages.pyrin = pyrin;

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            tools.publishVersion
          ];
        };
      }
    );
}
