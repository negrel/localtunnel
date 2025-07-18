{
  inputs = {
    flake-utils.url = "github:numtide/flake-utils";
  };

  outputs =
    {
      nixpkgs,
      flake-utils,
      ...
    }:
    let
      outputsWithoutSystem = { };
      outputsWithSystem = flake-utils.lib.eachDefaultSystem (
        system:
        let
          pkgs = import nixpkgs {
            inherit system;
          };
          lib = pkgs.lib;
        in
        {
          devShells = {
            default = pkgs.mkShell {
              buildInputs = with pkgs; [
                go
                gopls
              ];
            };
          };
          packages = {
            default = pkgs.buildGoModule {
              pname = "localtunnel";
              version = "0.1.0";

              src = ./.;
              vendorHash = "sha256-fB7SCab96NLc9jR95jfRLmsVA2IrlMckZ0t1E16AOZE=";

              postInstall = ''
                mv $out/bin/localtunnel $out/bin/lt
              '';
            };
          };
        }
      );
    in
    outputsWithSystem // outputsWithoutSystem;
}
