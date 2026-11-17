{
  perSystem =
    { pkgs, ... }:
    {
      # nixfmt on its own reads stdin, so `nix fmt` with no arguments does
      # nothing useful. nixfmt-tree wraps it in treefmt and walks the repository.
      formatter = pkgs.nixfmt-tree;
    };
}
