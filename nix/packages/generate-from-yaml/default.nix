{ lib, ...}:

{
  perSystem =
    { pkgs, ... }:
    let
      pkg = pkgs.buildGoModule {
        pname = "generate-from-yaml";
        version = "1.0.0";
        src = ./.;
        vendorHash = "sha256-ZqrQBD8aa0Mgn0JNqPSHcP2/Yc1H9wBJSwZRUZm+Ddw=";

        meta.mainProgram = "generate-from-yaml";
      };
    in
    {
      packages.generate-from-yaml = pkg;

      apps.generate-from-yaml.program = lib.getExe pkg;
    };
}
