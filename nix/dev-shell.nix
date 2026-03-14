{
  perSystem =
    { pkgs, ... }:
    {
      devShells.default = pkgs.mkShell {
        packages = with pkgs; [
          go
          gopls
          nixd
        ];
      };
    };
}
