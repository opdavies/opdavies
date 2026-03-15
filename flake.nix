{
  inputs = {
    flake-parts.url = "github:hercules-ci/flake-parts";
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs =
    inputs:
    inputs.flake-parts.lib.mkFlake { inherit inputs; } {
      imports = [
        ./nix/dev-shell.nix
        ./nix/formatter.nix
        ./nix/packages/generate-from-yaml
        ./nix/packages/update-readme
        ./nix/systems.nix
      ];
    };
}
