# Copyright 2024 Edgeless Systems GmbH
# SPDX-License-Identifier: AGPL-3.0-only

{ lib
, buildGoModule
, buildGoTest
, genpolicy-msft
, genpolicy ? genpolicy-msft
, contrast
, runtime-class-files
}:

let
  e2e = buildGoTest {
    inherit (contrast) version src proxyVendor vendorHash prePatch CGO_ENABLED;
    pname = "${contrast.pname}-e2e";

    tags = [ "e2e" ];

    ldflags = [
      "-s"
      "-X github.com/edgelesssys/contrast/internal/manifest.trustedMeasurement=${launchDigest}"
      "-X github.com/edgelesssys/contrast/cli/cmd.runtimeHandler=${runtimeHandler}"
      "-X github.com/edgelesssys/contrast/internal/kuberesource.runtimeHandler=${runtimeHandler}"
    ];

    subPackages = [ "e2e/getdents" "e2e/openssl" "e2e/servicemesh" "e2e/release" ];
  };

  launchDigest = builtins.readFile "${runtime-class-files}/launch-digest.hex";

  runtimeHandler = lib.removeSuffix "\n" (builtins.readFile "${runtime-class-files}/runtime-handler");

  packageOutputs = [ "coordinator" "initializer" "cli" ];
in

buildGoModule rec {
  pname = "contrast";
  version = builtins.readFile ../../../version.txt;

  outputs = packageOutputs ++ [ "out" ];

  # The source of the main module of this repo. We filter for Go files so that
  # changes in the other parts of this repo don't trigger a rebuild.
  src =
    let
      inherit (lib) fileset path hasSuffix;
      root = ../../../.;
    in
    fileset.toSource {
      inherit root;
      fileset = fileset.unions [
        (path.append root "go.mod")
        (path.append root "go.sum")
        (fileset.difference
          (fileset.fileFilter (file: hasSuffix ".go" file.name) root)
          (fileset.unions [
            (path.append root "service-mesh")
            (path.append root "node-installer")
          ]))
      ];
    };

  proxyVendor = true;
  vendorHash = "sha256-lakrQ7sWEaJ8Cjr15WtV0h3+YF69Qg9/VPJ1xZ7i5rI=";

  subPackages = packageOutputs ++ [ "internal/kuberesource/resourcegen" ];

  prePatch = ''
    install -D ${lib.getExe genpolicy} cli/cmd/assets/genpolicy
    install -D ${genpolicy.settings-dev}/genpolicy-settings.json cli/cmd/assets/genpolicy-settings.json
    install -D ${genpolicy.rules}/genpolicy-rules.rego cli/cmd/assets/genpolicy-rules.rego
  '';

  CGO_ENABLED = 0;
  ldflags = [
    "-s"
    "-w"
    "-X main.version=v${version}"
    "-X github.com/edgelesssys/contrast/internal/manifest.trustedMeasurement=${launchDigest}"
    "-X github.com/edgelesssys/contrast/cli/cmd.runtimeHandler=${runtimeHandler}"
    "-X github.com/edgelesssys/contrast/internal/kuberesource.runtimeHandler=${runtimeHandler}"
  ];

  preCheck = ''
    export CGO_ENABLED=1
  '';

  checkPhase = ''
    runHook preCheck
    go test -race ./...
    runHook postCheck
  '';

  postInstall = ''
    for sub in ${builtins.concatStringsSep " " packageOutputs}; do
      mkdir -p "''${!sub}/bin"
      mv "$out/bin/$sub" "''${!sub}/bin/$sub"
    done

    # rename the cli binary to contrast
    mv "$cli/bin/cli" "$cli/bin/contrast"
  '';

  passthru.e2e = e2e;

  meta.mainProgram = "contrast";
}
