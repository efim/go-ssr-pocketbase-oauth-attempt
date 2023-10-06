{
  description = "going to look at the pocketbase apis";

  outputs = { self, nixpkgs }: {
    devShell.x86_64-linux = let pkgs = nixpkgs.legacyPackages.x86_64-linux;
    in pkgs.mkShell {
      buildInputs = [
        pkgs.go
        pkgs.wgo # for restart of project
        pkgs.semgrep
        pkgs.gopls
        pkgs.nodePackages.tailwindcss
        pkgs.nodePackages.prettier
      ];

      shellHook = ''
        export GOPATH=$PWD/.go
        export PATH=$GOPATH/bin:$PATH
      '';
    };
  };
}
