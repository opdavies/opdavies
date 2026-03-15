{ lib, ... }:

{
  perSystem =
    { pkgs, ... }:
    let
      pkg = pkgs.buildGoModule {
        pname = "update-readme";
        version = "1.0.0";

        src = ./.;

        vendorHash = "sha256-ss7PrNrSuqsqmA/kfe1XpAW9dAeCAM9YlwsuQwn3OMA=";

        meta.mainProgram = "update-readme";
      };
    in
    {
      packages.update-readme = pkg;

      apps.update-readme.program = lib.getExe pkg;
    };
}
