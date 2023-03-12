#!/usr/bin/env bash

install_kind() {
  curl -Lo ./kind "https://kind.sigs.k8s.io/dl/v0.17.0/kind-$(uname)-amd64"
  chmod +x ./kind
  sudo mv ./kind /usr/local/bin/kind
}

install_tilt() {
  arch="linux"
  if [[ $(uname) == "Darwin" ]]; then
    arch="mac"
  fi
  curl -fsSL "https://github.com/tilt-dev/tilt/releases/download/v0.31.2/tilt.0.31.2.${arch}.x86_64.tar.gz" | tar -xzv tilt
  sudo mv ./tilt /usr/local/bin/tilt
}

install_kind
install_tilt
