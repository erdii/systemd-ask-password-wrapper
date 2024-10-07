#!/bin/bash
set -euxo pipefail

go build -o ./bin/systemd-ask-password-wrapper \
    ./cmd/systemd-ask-password-wrapper

sudo cp config/*.{path,service} /etc/systemd/system
sudo systemctl daemon-reload

sudo systemctl stop systemd-ask-password-wall.path \
                    systemd-ask-password-plymouth.path \
                    systemd-ask-password-wall.service \
                    systemd-ask-password-plymouth.service

sudo systemctl stop systemd-ask-password-wrapper.service
sudo systemctl restart systemd-ask-password-wrapper.path
