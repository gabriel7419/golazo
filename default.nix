{pkgs ? import <nixpkgs> {}, ...}:
pkgs.buildGoModule {
  pname = "golazo";
  version = "0.14.0";
  vendorHash = "sha256-+Nel552lu5VtLyP5Yv7CQs/SUT4S4+vZIxkqMeicbWg=";

  subPackages = ["."];

  src = builtins.path {
    path = ./.;
    name = "source";
  };
}
