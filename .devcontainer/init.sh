#!/usr/bin/env bash
set -euo pipefail

if [ -f backend/go.mod ]; then
  echo "Scaricando i moduli di GO..."
  (cd backend && go mod download)
fi

if [ -f frontend/package-lock.json ]; then
  echo "Installando le dipendenze frontend con npm ci...."
  (cd frontend && npm ci)
elif [ -f frontend/package.json ]; then #se manca lockfile
  echo "Installando le dipendenze fronted con npm..."
  (cd frontend && npm install)
fi
