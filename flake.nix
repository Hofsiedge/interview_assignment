{
  description = "An interview test for a Middle Go developer position";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
  };

  outputs =
    {
      nixpkgs,
      ...
    }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs { inherit system; };
    in
    {
      formatter.${system} = pkgs.nixfmt-rfc-style;

      devShells.${system}.default = pkgs.mkShell {
        buildInputs = with pkgs; [
          go_1_24
          golangci-lint
          gotools
          go-tools
          (go-migrate.overrideAttrs (oldAttrs: {
            tags = [ "postgres" ];
          }))
          redocly
        ];
        shellHook = ''
          export ROOT=$PWD

          # -- go setup --
          mkdir -p .nix-go
          # make go use a local directory
          export GOPATH=$ROOT/.nix-go
          # make executables available
          export PATH=$GOPATH/bin:$PATH
        '';
        name = "interview";
      };
    };
}
