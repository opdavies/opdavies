{
  perSystem =
    { lib, pkgs, ... }:
    let
      generate-from-yaml = pkgs.buildGoModule {
        pname = "generate-from-yaml";
        version = "1.0.0";
        src = ./generate-from-yaml;
        vendorHash = "sha256-ZqrQBD8aa0Mgn0JNqPSHcP2/Yc1H9wBJSwZRUZm+Ddw=";

        meta.mainProgram = "generate-from-yaml";
      };

      update-readme = pkgs.buildGoModule {
        pname = "update-readme";
        version = "1.0.0";
        src = ./update-readme;
        vendorHash = "sha256-ss7PrNrSuqsqmA/kfe1XpAW9dAeCAM9YlwsuQwn3OMA=";

        meta.mainProgram = "update-readme";
      };
    in
    {
      apps.generate-from-yaml = {
        type = "app";
        program = lib.getExe generate-from-yaml;
      };

      apps.update-readme = {
        type = "app";
        program = lib.getExe update-readme;
      };
    };
}
