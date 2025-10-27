# Empire

A microservices-based role-playing game management system featuring hierarchical roles, role-based access control, and weapon trading. The system is built with Go, using PostgreSQL for warrior management and MongoDB for weapon inventory, following CQRS pattern with Wire dependency injection.

## System Architecture

```mermaid
graph TB
    subgraph "Warrior Service (PostgreSQL)"
        WS[Warrior Service<br/>:8080]
        DB1[(PostgreSQL<br/>warriors)]
    end
    
    subgraph "Weapon Service (MongoDB)"
        WPS[Weapon Service<br/>:8081]
        DB2[(MongoDB<br/>weapons)]
    end
    
    subgraph "Shared Packages"
        AUTH[pkg/auth<br/>JWT]
        RBAC[pkg/rbac<br/>Access Control]
        VAL[pkg/validator<br/>Validation]
    end
    
    CLIENT[Client] -->|Login| WS
    CLIENT -->|Auth Token| WPS
    WS --> AUTH
    WPS --> AUTH
    WS --> RBAC
    WPS --> RBAC
    WS --> DB1
    WPS --> DB2
    WS --> AUTH
```

## Role Hierarchy

```mermaid
graph TD
    LE[Light Emperor]
    LK[Light King]
    K[Knight]
    A[Archer]
    M[Mage]
    
    DE[Dark Emperor]
    DK[Dark King]
    
    LE -->|Creates| LK
    LE -->|Creates| K
    LE -->|Creates| A
    LE -->|Creates| M
    
    LK -->|Creates| K
    LK -->|Creates| A
    LK -->|Creates| M
    
    DE -->|Creates| DK
    
    style LE fill:#FFD700
    style DE fill:#4B0082
    style LK fill:#FFA500
    style DK fill:#800080
```

## Authentication Flow

```mermaid
sequenceDiagram
    participant Client
    participant WarriorService as Warrior Service<br/>(:8080)
    participant Auth as pkg/auth
    participant PG as PostgreSQL
    
    Client->>WarriorService: POST /api/login<br/>{username, password}
    WarriorService->>PG: Find warrior by username
    PG-->>WarriorService: Warrior data
    WarriorService->>Auth: HashPassword + Compare
    Auth-->>WarriorService: Password verified
    WarriorService->>Auth: GenerateToken(userID, username, role)
    Auth-->>WarriorService: JWT Token
    WarriorService-->>Client: {token, warrior}
    
    Client->Validation Token: Use in Authorization: Bearer
```

## Warrior Creation Flow

```mermaid
sequenceDiagram
    participant User as Light Emperor/King
    participant WS as Warrior Service
    participant Auth as Auth Middleware
    participant Service as Service Layer
    participant PG as PostgreSQL
    
    User->>WS: POST /api/warriors<br/>+ JWT Token
    WS->>Auth: Validate Token
    Auth->>Auth: Extract Claims<br/>{userID, username, role}
    Auth-->>WS: User info
    
    alt Role is Light Emperor or Light King
        WS->>Service: CreateWarrior(command)
        Service->>Service: Validate role
        Service->>PG: Insert warrior
        PG-->>Service: Created warrior
        Service-->>WS: Success
        WS-->>User: 201 Created<br/>{warrior}
    else Role not authorized
        WS-->>User: 403 Forbidden
    end
```

## Weapon Purchase Flow

```mermaid
sequenceDiagram
    participant User as Any Warrior
    participant WPS as Weapon Service<br/>:8081
    participant Auth as JWT Auth
    participant Service as Weapon Service
    participant MongoDB as MongoDB
    
    User->>WPS: POST /api/weapons/buy<br/>+ JWT Token<br/>{weapon_id}
    WPS->>Auth: Validate Token
    Auth-->>WPS: {userID, role}
    
    WPS->>Service: BuyWeapon(command)
    Service->>MongoDB: Find weapon by ID
    MongoDB-->>Service: Weapon data
    
    alt CanBuy = true
        alt Not already owned
            Service->>MongoDB: Update weapon<br/>Add buyer to owned_by
            MongoDB-->>Service: Success
            Service-->>WPS: Success
            WPS-->>User: 200 OK
        else Already owned
            Service-->>WPS: Error
            WPS-->>User: 400 Already owned
        end
    else Role cannot buy
        Service-->>WPS: Error
        WPS-->>User: 403 Forbidden
    end
```

## RBAC Matrix

```mermaid
graph LR
    subgraph "Actions"
        CW[Create Warriors]
        CK[Create Kings]
        UW[Update Warriors]
        DW[Delete Warriors]
        CWEP[Create Weapons]
        BWC[Buy Common]
        BWR[Buy Rare]
        BWL[Buy Legendary]
    end
    
    subgraph "Light Emperor"
        LE[CW:✓ CK:✓ UW:✓ DW:✓<br/>CWEP:✓ BWC:✓ BWR:✓ BWL:✓]
    end
    
    subgraph "Light King"
        LK[CW:✓ CK:✗ UW:Self DW:✓<br/>CWEP:✓ BWC:✓ BWR:✓ BWL:✗]
    end
    
    subgraph "Knights/Archers/Mages"
        WAR[CW:✗ CK:✗ UW:Self DW:✗<br/>CWEP:✗ BWC:✓ BWR:✗ BWL:✗]
    end
    
    subgraph "Dark Emperor"
        DE[CW:✗ CK:✓ UW:Self DW:✗<br/>CWEP:✗ BWC:✗ BWR:✗ BWL:✗]
    end
    
    subgraph "Dark King"
        DK[CW:✗ CK:✗ UW:Self DW:✗<br/>CWEP:✗ BWC:✗ BWR:✗ BWL:✗]
    end
```

## Weapon Type Access

```mermaid
stateDiagram-v2
    [*] --> Common
    [*] --> Rare
    [*] --> Legendary
    
    Common --> KnightOwns: Knight/Archer/Mage
    Common --> KingOwns: Light King/Emperor
    
    Rare --> KingOwns: Light King/Emperor
    
    Legendary --> EmperorOwns: Light Emperor ONLY
    
    KnightOwns --> [*]
    KingOwns --> [*]
    EmperorOwns --> [*]
```

## Service Dependencies

```mermaid
graph TD
    subgraph "cmd/warrior"
        WM[main.go]
        WIRE[wire_gen.go]
    end
    
    subgraph "internal/warrior"
        WS[service.go]
        WH[handlers.go]
        CRUD[crud_handlers.go]
        KH[king_handlers.go]
        R[routes.go]
    end
    
    subgraph "cmd/weapon"
        WPM[main.go]
    end
    
    subgraph "internal/weapon"
        WPS[service.go]
        WPH[handlers.go]
        WR[routes.go]
    end
    
    subgraph "pkg/*"
        AUTH[pkg/auth]
        RBAC[pkg/rbac]
        VAL[pkg/validator]
    end
    
    WM --> WS
    WM --> WH
    WM --> CRUD
    WM --> KH
    WM --> R
    WM --> AUTH
    WM --> RBAC
    WM --> VAL
    
    WPM --> WPS
    WPM --> WPH
    WPM --> WR
    WPM --> AUTH
    WPM --> RBAC
    WPM --> VAL
    
    WH --> WS
    WPH --> WPS
```

## Quick Start

```bash
# Build warrior service
cd cmd/warrior && go run main.go

# Build weapon service  
cd cmd/weapon && go run main.go

# With Docker Compose
docker-compose up -d

# Services
# - Warrior: localhost:8080
# - Weapon: localhost:8081
# - PostgreSQL: localhost:5432
# - MongoDB: localhost:27017
```
