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

          ldflags = [
            "-X github.com/nanoteck137/pyrin/cmd.Version=${version}"
            "-X github.com/nanoteck137/pyrin/cmd.Commit=${self.dirtyRev or self.rev or "no-commit"}"
          ];

          vendorHash = "sha256-YStNcVhK9l1IF1F5OWHubEzkqZempK71HAUntPXeGak=";
        };
      in
      {
        packages.default = pyrin;
        packages.pyrin = pyrin;

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
          ];
        };
      }
    );
}
