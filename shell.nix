{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  buildInputs = [
    pkgs.go_1_20
    pkgs.delve
    pkgs.gdlv
  ];
}
