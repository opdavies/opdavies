{
  perSystem =
    {
      lib,
      pkgs,
      self',
      ...
    }:
    {
      devShells.default = pkgs.mkShell {
        inputsFrom = lib.attrValues self'.packages;

        packages =
          with pkgs;
          [
            gopls
            nixd
          ]
          + lib.attrValues self'.packages;
      };
    };
}
