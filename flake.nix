{
  inputs = {
    flake-parts.url = "github:hercules-ci/flake-parts";
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    inputs:
    inputs.flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [
        ./flake-parts/dev-shell.nix
        ./flake-parts/formatter.nix
        ./flake-parts/packages/generate-from-yaml
        ./flake-parts/packages/update-readme
        ./flake-parts/systems.nix
      ];
    };
}
