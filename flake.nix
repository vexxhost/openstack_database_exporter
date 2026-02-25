{
  inputs = {
    nixpkgs = {
      url = "github:NixOS/nixpkgs/nixpkgs-unstable";
    };

    flake-utils = {
      url = "github:numtide/flake-utils";
    };

    nix2container = {
      url = "github:nlewo/nix2container";
      inputs.nixpkgs.follows = "nixpkgs";
      inputs.flake-utils.follows = "flake-utils";
    };
  };

  outputs =
    {
      self,
      nixpkgs,
      flake-utils,
      nix2container,
    }:
    flake-utils.lib.eachDefaultSystem (
      system:
      let
        name = "openstack-database-exporter";

        pkgs = import nixpkgs {
          inherit system;
        };
        nix2containerPkgs = nix2container.packages.${system};
      in
      rec {
        apps.copyDockerImage = {
          type = "app";
          program = builtins.toString (pkgs.writeShellScript "copyDockerImage" ''
            IFS=$'\n' # iterate over newlines
            set -x # echo on
            for DOCKER_TAG in $DOCKER_METADATA_OUTPUT_TAGS; do
              ${pkgs.lib.getExe self.packages.${system}.dockerImage.copyTo} "docker://$DOCKER_TAG"
            done
          '');
        };

        packages = rec {
          openstack-database-exporter = pkgs.buildGoModule {
            pname = name;
            version = "0.1.0";
            src = ./.;
            vendorHash = null;
            subPackages = [ "cmd/openstack-database-exporter" ];

            CGO_ENABLED = 0;

            meta = {
              description = "OpenStack Database Exporter for Prometheus";
              mainProgram = "openstack-database-exporter";
            };
          };
          default = packages.openstack-database-exporter;

          dockerImage = nix2containerPkgs.nix2container.buildImage {
            name = "ghcr.io/vexxhost/openstack-database-exporter";
            maxLayers = 64;
            copyToRoot = with pkgs.dockerTools; [
              caCertificates
            ];

            config = {
              entrypoint = [ "${default}/bin/openstack-database-exporter" ];
              ExposedPorts = {
                "9180/tcp" = { };
              };
            };
          };
        };

        devShell = pkgs.mkShell {
          buildInputs = with pkgs; [
            go
            golangci-lint
            mariadb
            sqlc
          ];
        };
      }
    );
}
