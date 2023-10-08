{
  description = "going to look at the pocketbase apis";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = nixpkgs.legacyPackages.${system};
        pname = "auth-pocketbase-attempt";
        version = "0.0.1";
      in rec {
        devShells.default = pkgs.mkShell {
          buildInputs = [
            pkgs.go
            pkgs.wgo # for restart of project
            pkgs.semgrep
            pkgs.gopls
            pkgs.nodePackages.tailwindcss
            pkgs.nodePackages.prettier
            pkgs.gnumake
          ];

          shellHook = ''
            export GOPATH=$PWD/.go
            export PATH=$GOPATH/bin:$PATH
          '';
        };
        packages = rec {
          auth-pocketbase-attempt = pkgs.buildGoModule {
            inherit pname version;
            src = pkgs.nix-gitignore.gitignoreSource [ ] ./.;
            vendorHash = "sha256-7B5EkrLpL+P5wipQG5a12hrvXQn/UpYAjrz/DuHmSUQ="; # set to "" when get dependencies in go.mod

            # Adding the Tailwind build step to preBuild
            preBuild = ''
              ${pkgs.nodePackages.tailwindcss}/bin/tailwindcss -i pages/input.css -o pages/static/public/out.css
            '';
          };
          default = auth-pocketbase-attempt;
        };
      });
}
