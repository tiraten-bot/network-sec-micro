#!/bin/bash

# Run script for all services (development mode)

set -e

echo "🚀 Starting all services..."
echo ""
echo "💡 Tip: Use separate terminals for each service"
echo "   Warrior: bash scripts/warrior-run.sh"
echo "   Weapon:  bash scripts/weapon-run.sh"
echo "   Coin:    bash scripts/coin-run.sh"
echo "   Enemy:   bash scripts/enemy-run.sh"
echo "   Dragon:  bash scripts/dragon-run.sh"
echo ""
echo "   Or use Docker: docker-compose up"
echo ""
echo "Starting Warrior service on :8080..."

# Start Warrior service (blocking)
bash scripts/warrior-run.sh
