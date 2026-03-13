#!/bin/bash
set -e

# Reload systemd and enable/start the service
systemctl daemon-reload
systemctl enable godshell.service
systemctl restart godshell.service

echo "Godshell daemon has been installed and started."
