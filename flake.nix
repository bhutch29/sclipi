{
  description = "SCPI CLI";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs = { self, nixpkgs, flake-utils, ... }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs {
          inherit system;
        };
      in
      {
        packages = {
          default = pkgs.buildGoModule {
            pname = "sclipi";
            version = "0.4.0";
            src = ./.;
            vendorHash = "sha256-8e2Ae3dqnhkl4fMPm9lJn7gDyTTnQGR7eXJn4YyxcDs=";
            meta = {
              description = "SCPI CLI";
            };
          };

          nixosModule = { config, lib, pkgs, ... }:
            let
              cfg = config.programs.sclipi;
              sclipiPkg = self.packages.${system}.default;
            in
            {
              options.programs.sclipi = {
                enable = lib.mkEnableOption "sclipi";
              };
              config = lib.mkIf cfg.enable {
                environment.systemPackages = [ sclipiPkg ];
              };
            };
        };

        devShells.default = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            gopls
            gotools
            nodejs
            just
          ];
        };
      }
    );
}
