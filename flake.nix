{
  inputs = {
    flake-parts.url = "github:hercules-ci/flake-parts";
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    inputs:
    inputs.flake-parts.lib.mkFlake { inherit inputs; } {

      imports = [
        ./nix/formatter.nix
        ./nix/systems.nix
      ];

      perSystem =
        { lib, pkgs, ... }:
        let
          generate-from-yaml = pkgs.buildGoModule {
            name = "generate-from-yaml";

            runtimeInputs = with pkgs; [ go ];

            src = ./nix/generate-from-yaml;

            vendorHash = "sha256-ZqrQBD8aa0Mgn0JNqPSHcP2/Yc1H9wBJSwZRUZm+Ddw=";

            meta.mainProgram = "generate-from-yaml";
          };

          update-readme = pkgs.buildGoModule {
            name = "update-readme";

            runtimeInputs = with pkgs; [ go ];

            src = ./nix/update-readme;

            vendorHash = "sha256-ss7PrNrSuqsqmA/kfe1XpAW9dAeCAM9YlwsuQwn3OMA=";

            meta.mainProgram = "update-readme";
          };
        in
        {

          devShells.default = pkgs.mkShell {
            packages = with pkgs; [
              go
              gopls
              nixd
            ];
          };

          apps.generate-from-yaml = {
            type = "app";
            program = lib.getExe generate-from-yaml;
          };

          apps.update-readme = {
            type = "app";
            program = lib.getExe update-readme;
          };
        };
    };
}
