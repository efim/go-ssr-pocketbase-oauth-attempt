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
            vendorHash =
              "sha256-7B5EkrLpL+P5wipQG5a12hrvXQn/UpYAjrz/DuHmSUQ="; # set to "" when get dependencies in go.mod

            # Adding the Tailwind build step to preBuild
            preBuild = ''
              ${pkgs.nodePackages.tailwindcss}/bin/tailwindcss -i pages/input.css -o pages/static/public/out.css
            '';
          };
          default = auth-pocketbase-attempt;
        };
        nixosModules.auth-pocketbase-attempt = { config, pkgs, ... }:
          let
            cfg = config.services.${pname};
            lib = pkgs.lib;
            shortName = "pb-auth-example-group";
          in {
            options.services.${pname} = {
              enable = lib.mkEnableOption
                "Enable simple ssr oauth example build on pocketbase";
              port = lib.mkOption {
                type = lib.types.int;
                default = 8090;
                description =
                  "Port to listen on. Use 443 for tls when no nginx, usual plaintext is 8090.";
              };
              host = lib.mkOption {
                type = lib.types.str;
                default = "127.0.0.1";
                description = "Host to bind to.";
              };
              useNginx = lib.mkOption {
                type = lib.types.bool;
                default = true;
                description = "Whether to use Nginx to proxy requests.";
              };
              usePbTls = lib.mkOption {
                type = lib.types.bool;
                default = false;
                description =
                  "Whether pocketbase should serve on https and issue own certs. Main case for true - when not under nginx";
              };
            };
            config = lib.mkIf cfg.enable {
              users.groups."${shortName}-group" = { };
              users.users."${shortName}-user" = {
                isSystemUser = true;
                group = "${shortName}-group";
              };
              systemd.services.${shortName} = let
                protocol = if cfg.usePbTls then "https" else "http";
                serverHost = if cfg.useNginx then "127.0.0.1" else cfg.host;
                servedAddress = "${protocol}://${serverHost}:${cfg.port}";
              in {
                description = "Exercise app ${pname}";
                wantedBy = [ "multi-user.target" ];
                after = [ "network.target" ];
                startLimitIntervalSec = 30;
                startLimitBurst = 10;
                serviceConfig = {
                  ExecStart =
                    "${packages.auth-pocketbase-attempt}/bin/${pname} serve ${servedAddress} --dir=/home/${
                      config.users.users."${shortName}-user"
                    }";
                  Restart = "on-failure";
                  User = "${shortName}-user";
                  Group = "${shortName}-group";
                };
              };
            };
          };
      });
}
