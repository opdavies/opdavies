{ lib, ... }:

{
  perSystem =
    { pkgs, ... }:
    let
      pkg = pkgs.buildGoModule {
        pname = "update-readme";
        version = "1.0.0";

        src = ./.;

        # No dependencies: the program uses only the standard library.
        vendorHash = null;

        meta = {
          description = "Update the generated sections of README.md from oliverdavies.uk";
          mainProgram = "update-readme";
        };
      };
    in
    {
      packages.update-readme = pkg;

      apps.update-readme = {
        program = lib.getExe pkg;
        meta.description = "Update the generated sections of README.md from oliverdavies.uk";
      };
    };
}
