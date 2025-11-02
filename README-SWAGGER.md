# API Documentation

Bu proje tüm HTTP servisleri için Swagger/OpenAPI dokümantasyonu içerir.

## Servisler ve Swagger URL'leri

### Warrior Service
- **Port**: 8080
- **Swagger UI**: http://localhost:8080/swagger/index.html
- **Base Path**: `/api`

### Weapon Service
- **Port**: 8081
- **Swagger UI**: http://localhost:8081/swagger/index.html
- **Base Path**: `/api`

### Dragon Service
- **Port**: 8084
- **Swagger UI**: http://localhost:8084/swagger/index.html
- **Base Path**: `/api/v1`

### Coin Service
- **Port**: 50051 (gRPC)
- **Dokümantasyon**: gRPC servis olduğu için protobuf dosyalarından yararlanılabilir
- **Proto Files**: `api/proto/coin/coin.proto`

## Swagger Dokümanlarını Generate Etme

Swagger dokümanlarını yeniden oluşturmak için:

```bash
./scripts/swagger-gen.sh
```

Veya manuel olarak her servis için:

```bash
# Warrior
cd cmd/warrior
swag init --parseDependency --parseInternal

# Weapon
cd cmd/weapon
swag init --parseDependency --parseInternal

# Dragon
cd cmd/dragon
swag init --parseDependency --parseInternal
```

## Authentication

Çoğu endpoint JWT token ile korunmaktadır. Swagger UI'da token'ı şu şekilde kullanabilirsiniz:

1. Swagger UI'ın sağ üstündeki "Authorize" butonuna tıklayın
2. `Bearer <your-token>` formatında token'ı girin
3. Örnek: `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...`

## API Gateway

Fiber Gateway şu anda declarative routing kullanmaktadır. Gateway üzerinden erişilen servislerin Swagger dokümantasyonları kendi servis portlarından erişilebilir.

Gateway endpoint'leri için detaylı konfigürasyon `fiber-gateway/config.example.json` dosyasında bulunabilir.

