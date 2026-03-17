{
  perSystem =
    { pkgs, self', ... }:
    {
      devShells.default = pkgs.mkShell {
        inputsFrom = [
          self'.packages.generate-from-yaml
          self'.packages.update-readme
        ];

        packages = with pkgs; [
          gopls
          nixd

          self'.packages.generate-from-yaml
          self'.packages.update-readme
        ];
      };
    };
}
