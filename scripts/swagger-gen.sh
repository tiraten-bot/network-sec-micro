#!/bin/bash

# Generate Swagger docs for all services
echo "Generating Swagger documentation..."

# Install swag if not already installed
if ! command -v swag &> /dev/null; then
    echo "Installing swag..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi

# Warrior service
echo "Generating docs for warrior service..."
cd cmd/warrior
swag init --parseDependency --parseInternal
cd ../..

# Weapon service
echo "Generating docs for weapon service..."
cd cmd/weapon
swag init --parseDependency --parseInternal
cd ../..

# Dragon service
echo "Generating docs for dragon service..."
cd cmd/dragon
swag init --parseDependency --parseInternal
cd ../..

# Battle service
echo "Generating docs for battle service..."
cd cmd/battle
swag init --parseDependency --parseInternal
cd ../..

echo "Swagger documentation generated successfully!"
echo ""
echo "Access Swagger UI at:"
echo "  Warrior: http://localhost:8080/swagger/index.html"
echo "  Weapon:  http://localhost:8081/swagger/index.html"
echo "  Dragon:  http://localhost:8084/swagger/index.html"
echo "  Battle:  http://localhost:8085/swagger/index.html"

