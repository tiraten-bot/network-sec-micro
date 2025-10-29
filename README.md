# Empire

A comprehensive microservices-based game system featuring warriors, weapons, coins, enemies, and dragons with role-based access control, gRPC communication, and event-driven architecture using Kafka.

## Architecture Overview

```mermaid
graph TB
    subgraph "Client Layer"
        WEB[Web Client]
        API[API Client]
    end
    
    subgraph "API Gateway Layer"
        LB[Load Balancer]
    end
    
    subgraph "Microservices Layer"
        W[Warrior Service<br/>HTTP :8080]
        WP[Weapon Service<br/>HTTP :8081]
        C[Coin Service<br/>gRPC :50051]
        E[Enemy Service<br/>HTTP :8083]
        D[Dragon Service<br/>HTTP :8084]
    end
    
    subgraph "Data Layer"
        PG[(PostgreSQL<br/>Warrior Data)]
        MG[(MongoDB<br/>Weapon/Enemy/Dragon)]
        MY[(MySQL<br/>Coin Transactions)]
    end
    
    subgraph "Event Layer"
        K[Kafka<br/>Event Streaming]
        Z[Zookeeper<br/>Kafka Coordination]
    end
    
    WEB --> LB
    API --> LB
    LB --> W
    LB --> WP
    LB --> E
    LB --> D
    
    W -.->|gRPC| C
    E -.->|gRPC| W
    D -.->|gRPC| W
    
    W --> PG
    WP --> MG
    E --> MG
    D --> MG
    C --> MY
    
    WP -->|Events| K
    E -->|Events| K
    D -->|Events| K
    C -->|Consume| K
    WP -->|Consume| K
    
    K --> Z
    
    style WEB fill:#FFD700,stroke:#000000,stroke-width:3px
    style API fill:#FFD700,stroke:#000000,stroke-width:3px
    style LB fill:#FF8C00,stroke:#8B0000,stroke-width:3px
    style W fill:#DC143C,stroke:#8B0000,stroke-width:3px
    style WP fill:#DC143C,stroke:#8B0000,stroke-width:3px
    style C fill:#DC143C,stroke:#8B0000,stroke-width:3px
    style E fill:#DC143C,stroke:#8B0000,stroke-width:3px
    style D fill:#DC143C,stroke:#8B0000,stroke-width:3px
    style PG fill:#FF8C00,stroke:#8B0000,stroke-width:3px
    style MG fill:#FF8C00,stroke:#8B0000,stroke-width:3px
    style MY fill:#FF8C00,stroke:#8B0000,stroke-width:3px
    style K fill:#FFD700,stroke:#000000,stroke-width:3px
    style Z fill:#FF8C00,stroke:#8B0000,stroke-width:3px
```

## Service Communication Flow

```mermaid
sequenceDiagram
    participant Client
    participant Warrior as Warrior Service
    participant Weapon as Weapon Service
    participant Coin as Coin Service
    participant Enemy as Enemy Service
    participant Dragon as Dragon Service
    participant Kafka as Kafka Events
    
    Note over Client,Dragon: Warrior Weapon Purchase Flow
    Client->>Warrior: Login & Get Token
    Warrior->>Weapon: Validate Token (gRPC)
    Weapon->>Coin: Deduct Coins (gRPC)
    Coin-->>Weapon: Payment Confirmed
    Weapon->>Kafka: Publish Purchase Event
    Weapon-->>Client: Weapon Purchased
    
    Note over Client,Dragon: Enemy Attack Flow
    Client->>Enemy: Create Goblin (Dark Emperor)
    Enemy->>Warrior: Get Warrior Info (gRPC)
    Enemy->>Enemy: Attack Warrior
    Enemy->>Kafka: Publish Attack Event
    Coin->>Kafka: Consume Attack Event
    Coin->>Coin: Deduct Warrior Coins
    
    Note over Client,Dragon: Dragon Battle Flow
    Client->>Dragon: Create Dragon (Dark Emperor)
    Client->>Dragon: Attack Dragon (Light King/Emperor)
    Dragon->>Warrior: Get Warrior Info (gRPC)
    Dragon->>Dragon: Calculate Damage
    Dragon->>Kafka: Publish Death Event (if killed)
    Weapon->>Kafka: Consume Death Event
    Weapon->>Weapon: Add Loot Weapon
```

## Role-Based Access Control (RBAC)

```mermaid
graph TD
    subgraph "Light Side"
        LK[Light King<br/>Can kill dragons<br/>Can view all balances]
        LE[Light Emperor<br/>Can kill dragons<br/>Can view all balances]
        LW[Light Warrior<br/>Can buy weapons<br/>Can view own balance]
    end
    
    subgraph "Dark Side"
        DK[Dark King<br/>Can create enemies<br/>Can view all balances]
        DE[Dark Emperor<br/>Can create enemies & dragons<br/>Can view all balances]
        DW[Dark Warrior<br/>Can buy weapons<br/>Can view own balance]
    end
    
    subgraph "Actions"
        A1[Create Warriors]
        A2[Buy Weapons]
        A3[Create Enemies]
        A4[Create Dragons]
        A5[Kill Dragons]
        A6[View Balances]
    end
    
    LK --> A5
    LK --> A6
    LE --> A5
    LE --> A6
    LW --> A2
    LW --> A6
    
    DK --> A3
    DK --> A6
    DE --> A3
    DE --> A4
    DE --> A6
    DW --> A2
    DW --> A6
    
    style LK fill:#FFD700,stroke:#000000,stroke-width:3px
    style LE fill:#FFD700,stroke:#000000,stroke-width:3px
    style LW fill:#FF8C00,stroke:#8B0000,stroke-width:3px
    style DK fill:#DC143C,stroke:#8B0000,stroke-width:3px
    style DE fill:#DC143C,stroke:#8B0000,stroke-width:3px
    style DW fill:#FF8C00,stroke:#8B0000,stroke-width:3px
    style A1 fill:#FFD700,stroke:#000000,stroke-width:3px
    style A2 fill:#FF8C00,stroke:#8B0000,stroke-width:3px
    style A3 fill:#DC143C,stroke:#8B0000,stroke-width:3px
    style A4 fill:#DC143C,stroke:#8B0000,stroke-width:3px
    style A5 fill:#FFD700,stroke:#000000,stroke-width:3px
    style A6 fill:#FF8C00,stroke:#8B0000,stroke-width:3px
```

## Database Architecture

```mermaid
erDiagram
    WARRIOR {
        uint id PK
        string username UK
        string email UK
        string password
        string role
        int coin_balance
        int total_power
        int weapon_count
        timestamp created_at
        timestamp updated_at
    }
    
    WEAPON {
        string id PK
        string name
        string type
        int attack_power
        int price
        string created_by
        array owned_by
        timestamp created_at
        timestamp updated_at
    }
    
    TRANSACTION {
        uint id PK
        uint warrior_id FK
        int64 amount
        string transaction_type
        string reason
        int64 balance_before
        int64 balance_after
        timestamp created_at
    }
    
    ENEMY {
        objectid id PK
        string name
        string type
        int level
        int health
        int attack_power
        string created_by
        timestamp created_at
        timestamp updated_at
    }
    
    DRAGON {
        objectid id PK
        string name
        string type
        int level
        int health
        int max_health
        int attack_power
        int defense
        string created_by
        boolean is_alive
        string killed_by
        timestamp killed_at
        timestamp created_at
        timestamp updated_at
    }
    
    WARRIOR ||--o{ TRANSACTION : "has"
    WARRIOR ||--o{ WEAPON : "owns"
```

## Event-Driven Architecture

```mermaid
graph LR
    subgraph "Event Producers"
        WP[Weapon Service]
        E[Enemy Service]
        D[Dragon Service]
    end
    
    subgraph "Kafka Topics"
        T1[weapon.purchase]
        T2[enemy.attack]
        T3[dragon.death]
    end
    
    subgraph "Event Consumers"
        C[Coin Service]
        WP2[Weapon Service]
    end
    
    WP -->|WeaponPurchaseEvent| T1
    E -->|EnemyAttackEvent| T2
    D -->|DragonDeathEvent| T3
    
    T1 --> C
    T2 --> C
    T3 --> WP2
    
    C -->|Deduct Coins| C
    WP2 -->|Add Loot Weapon| WP2
    
    style WP fill:#DC143C,stroke:#8B0000,stroke-width:3px
    style E fill:#DC143C,stroke:#8B0000,stroke-width:3px
    style D fill:#DC143C,stroke:#8B0000,stroke-width:3px
    style T1 fill:#FFD700,stroke:#000000,stroke-width:3px
    style T2 fill:#FFD700,stroke:#000000,stroke-width:3px
    style T3 fill:#FFD700,stroke:#000000,stroke-width:3px
    style C fill:#FF8C00,stroke:#8B0000,stroke-width:3px
    style WP2 fill:#FF8C00,stroke:#8B0000,stroke-width:3px
```

## Gateway Routing and Resilience

```mermaid
%%{init: {
  'theme': 'base',
  'themeVariables': {
    'primaryColor': '#0b3d91',
    'primaryTextColor': '#ffffff',
    'primaryBorderColor': '#001a4d',
    'lineColor': '#001a4d',
    'tertiaryColor': '#0d56b3',
    'clusterBkg': '#0b3d91',
    'clusterBorder': '#001a4d'
  }
}}%%
flowchart LR
    subgraph GW["Fiber API Gateway"]
        RL[Rate Limiter - Token Bucket/Redis]
        CB[Circuit Breaker]
        LB[Load Balancer - RR/Least-Conn]
        OD[Outlier Detection]
        TR[Transforms - headers/query/rewrite]
        AGG[Aggregates - fan-out/fan-in]
        CACHE[Response Cache - ETag/TTL]
    end

    C[(Client)] --> RL --> CB --> TR --> LB --> OD -->|HTTP/WS/gRPC h2c| UP[(Upstreams)]
    TR --> AGG --> CACHE

    style GW fill:#0b3d91,stroke:#001a4d,stroke-width:2px,color:#ffffff
    style RL fill:#133e7c,stroke:#001a4d,color:#ffffff
    style CB fill:#133e7c,stroke:#001a4d,color:#ffffff
    style LB fill:#133e7c,stroke:#001a4d,color:#ffffff
    style OD fill:#133e7c,stroke:#001a4d,color:#ffffff
    style TR fill:#0d56b3,stroke:#001a4d,color:#ffffff
    style AGG fill:#0d56b3,stroke:#001a4d,color:#ffffff
    style CACHE fill:#0d56b3,stroke:#001a4d,color:#ffffff
    style C fill:#08315c,stroke:#001a4d,color:#ffffff
    style UP fill:#08315c,stroke:#001a4d,color:#ffffff
```

## Docker vs Kubernetes Topology

```mermaid
%%{init: {
  'theme': 'base',
  'themeVariables': {
    'primaryColor': '#0b3d91',
    'primaryTextColor': '#ffffff',
    'primaryBorderColor': '#001a4d',
    'lineColor': '#001a4d',
    'tertiaryColor': '#0d56b3',
    'clusterBkg': '#0b3d91',
    'clusterBorder': '#001a4d'
  }
}}%%
graph TB
    subgraph "Docker Compose"
        DCGW[Gateway:8090]
        DCW[Warrior:8080]
        DCWP[Weapon:8081]
        DCE[Enemy:8083]
        DCD[Dragon:8084]
        DCC[Coin:50051]
        DCK[Kafka:9092]
        DCR[Redis:6379]
        DCPG[Postgres]
        DCMG[Mongo]
        DCMY[MySQL]
    end

    subgraph "Kubernetes (Helm)"
        KSGW[Deployment fiber-gateway]
        KSW[Deployment warrior]
        KSWP[Deployment weapon]
        KSE[Deployment enemy]
        KSD[Deployment dragon]
        KSC[Deployment coin]
        KSR[bitnami/redis]
        KSK[bitnami/kafka + zookeeper]
        KSPG[bitnami/postgresql]
        KSMG[bitnami/mongodb]
        KSMY[bitnami/mysql]
        ING[Ingress]
    end

    DCGW --- DCW & DCWP & DCE & DCD
    DCK --- DCWP & DCE & DCD & DCC
    DCR --- DCGW
    DCPG --- DCW
    DCMG --- DCWP & DCE & DCD
    DCMY --- DCC

    ING --- KSGW
    KSGW --- KSW & KSWP & KSE & KSD
    KSK --- KSWP & KSE & KSD & KSC
    KSR --- KSGW
    KSPG --- KSW
    KSMG --- KSWP & KSE & KSD
    KSMY --- KSC

    style DCGW fill:#0b3d91,stroke:#001a4d,color:#ffffff
    style DCW fill:#0d56b3,stroke:#001a4d,color:#ffffff
    style DCWP fill:#0d56b3,stroke:#001a4d,color:#ffffff
    style DCE fill:#0d56b3,stroke:#001a4d,color:#ffffff
    style DCD fill:#0d56b3,stroke:#001a4d,color:#ffffff
    style DCC fill:#0d56b3,stroke:#001a4d,color:#ffffff
    style DCK fill:#133e7c,stroke:#001a4d,color:#ffffff
    style DCR fill:#133e7c,stroke:#001a4d,color:#ffffff
    style DCPG fill:#133e7c,stroke:#001a4d,color:#ffffff
    style DCMG fill:#133e7c,stroke:#001a4d,color:#ffffff
    style DCMY fill:#133e7c,stroke:#001a4d,color:#ffffff

    style KSGW fill:#0b3d91,stroke:#001a4d,color:#ffffff
    style KSW fill:#0d56b3,stroke:#001a4d,color:#ffffff
    style KSWP fill:#0d56b3,stroke:#001a4d,color:#ffffff
    style KSE fill:#0d56b3,stroke:#001a4d,color:#ffffff
    style KSD fill:#0d56b3,stroke:#001a4d,color:#ffffff
    style KSC fill:#0d56b3,stroke:#001a4d,color:#ffffff
    style KSK fill:#133e7c,stroke:#001a4d,color:#ffffff
    style KSR fill:#133e7c,stroke:#001a4d,color:#ffffff
    style KSPG fill:#133e7c,stroke:#001a4d,color:#ffffff
    style KSMG fill:#133e7c,stroke:#001a4d,color:#ffffff
    style KSMY fill:#133e7c,stroke:#001a4d,color:#ffffff
    style ING fill:#08315c,stroke:#001a4d,color:#ffffff
```

## Kafka Topics (Extended)

```mermaid
graph LR
    subgraph Topics
        T1[weapon.purchase]
        T2[enemy.attack]
        T3[dragon.death]
        T4[enemy.destroyed]
    end

    subgraph Producers
        P1[Weapon]
        P2[Enemy]
        P3[Dragon]
    end

    subgraph Consumers
        C1[Coin]
        C2[Weapon]
        C3[Warrior]
    end

    P1 --> T1
    P2 --> T2
    P3 --> T3
    P2 --> T4

    T1 --> C1
    T2 --> C1
    T3 --> C2
    T4 --> C3

    style T1 fill:#0b3d91,stroke:#001a4d,color:#ffffff
    style T2 fill:#0b3d91,stroke:#001a4d,color:#ffffff
    style T3 fill:#0b3d91,stroke:#001a4d,color:#ffffff
    style T4 fill:#0b3d91,stroke:#001a4d,color:#ffffff
    style P1 fill:#0d56b3,stroke:#001a4d,color:#ffffff
    style P2 fill:#0d56b3,stroke:#001a4d,color:#ffffff
    style P3 fill:#0d56b3,stroke:#001a4d,color:#ffffff
    style C1 fill:#133e7c,stroke:#001a4d,color:#ffffff
    style C2 fill:#133e7c,stroke:#001a4d,color:#ffffff
    style C3 fill:#133e7c,stroke:#001a4d,color:#ffffff
```

## Deployment Pipeline (Local → Docker → Helm)

```mermaid
sequenceDiagram
    participant Dev as Developer
    participant Local as Local (scripts)
    participant DC as Docker Compose
    participant Helm as Helmfile (Infra)
    participant Chart as Helm (App)

    Dev->>Local: ./scripts/*.sh
    Note right of Local: build/run/test
    Local-->>DC: docker-build.sh / docker-run.sh
    Dev->>Helm: helm-apply.sh (infra deps)
    Dev->>Chart: helm-app-apply.sh (app chart)
    Chart-->>Dev: Ingress URL
```

## Service Dependencies

```mermaid
graph TD
    subgraph "Core Services"
        W[Warrior Service<br/>Authentication & User Management]
    end
    
    subgraph "Game Services"
        WP[Weapon Service<br/>Weapon Management]
        C[Coin Service<br/>Transaction Management]
        E[Enemy Service<br/>Enemy Management]
        D[Dragon Service<br/>Dragon Management]
    end
    
    subgraph "Infrastructure"
        K[Kafka<br/>Event Streaming]
        PG[PostgreSQL<br/>Warrior Data]
        MG[MongoDB<br/>Game Data]
        MY[MySQL<br/>Transaction Data]
    end
    
    W --> PG
    WP --> MG
    E --> MG
    D --> MG
    C --> MY
    
    WP -.->|gRPC| W
    E -.->|gRPC| W
    D -.->|gRPC| W
    WP -.->|gRPC| C
    
    WP -->|Events| K
    E -->|Events| K
    D -->|Events| K
    C -->|Consume| K
    WP -->|Consume| K
    
    style W fill:#FFD700,stroke:#000000,stroke-width:3px
    style WP fill:#DC143C,stroke:#8B0000,stroke-width:3px
    style C fill:#DC143C,stroke:#8B0000,stroke-width:3px
    style E fill:#DC143C,stroke:#8B0000,stroke-width:3px
    style D fill:#DC143C,stroke:#8B0000,stroke-width:3px
    style K fill:#FFD700,stroke:#000000,stroke-width:3px
    style PG fill:#FF8C00,stroke:#8B0000,stroke-width:3px
    style MG fill:#FF8C00,stroke:#8B0000,stroke-width:3px
    style MY fill:#FF8C00,stroke:#8B0000,stroke-width:3px
```

## API Endpoints Overview

```mermaid
graph TB
    subgraph "Warrior Service :8080"
        W1[POST /api/v1/warriors/register]
        W2[POST /api/v1/warriors/login]
        W3[GET /api/v1/warriors/profile]
        W4[PUT /api/v1/warriors/profile]
    end
    
    subgraph "Weapon Service :8081"
        WP1[GET /api/v1/weapons]
        WP2[POST /api/v1/weapons]
        WP3[POST /api/v1/weapons/:id/buy]
        WP4[GET /api/v1/weapons/my-weapons]
    end
    
    subgraph "Coin Service :50051"
        C1[GetBalance]
        C2[DeductCoins]
        C3[AddCoins]
        C4[TransferCoins]
        C5[GetTransactionHistory]
    end
    
    subgraph "Enemy Service :8083"
        E1[POST /api/v1/enemies]
        E2[POST /api/v1/enemies/:id/attack]
        E3[GET /api/v1/enemies]
        E4[GET /api/v1/enemies/type/:type]
    end
    
    subgraph "Dragon Service :8084"
        D1[POST /api/v1/dragons]
        D2[POST /api/v1/dragons/:id/attack]
        D3[GET /api/v1/dragons/:id]
        D4[GET /api/v1/dragons/type/:type]
    end
    
    style W1 fill:#FF8C00,stroke:#8B0000,stroke-width:3px
    style W2 fill:#FF8C00,stroke:#8B0000,stroke-width:3px
    style W3 fill:#FF8C00,stroke:#8B0000,stroke-width:3px
    style W4 fill:#FF8C00,stroke:#8B0000,stroke-width:3px
    style WP1 fill:#DC143C,stroke:#8B0000,stroke-width:3px
    style WP2 fill:#DC143C,stroke:#8B0000,stroke-width:3px
    style WP3 fill:#DC143C,stroke:#8B0000,stroke-width:3px
    style WP4 fill:#DC143C,stroke:#8B0000,stroke-width:3px
    style C1 fill:#FFD700,stroke:#000000,stroke-width:3px
    style C2 fill:#FFD700,stroke:#000000,stroke-width:3px
    style C3 fill:#FFD700,stroke:#000000,stroke-width:3px
    style C4 fill:#FFD700,stroke:#000000,stroke-width:3px
    style C5 fill:#FFD700,stroke:#000000,stroke-width:3px
    style E1 fill:#DC143C,stroke:#8B0000,stroke-width:3px
    style E2 fill:#DC143C,stroke:#8B0000,stroke-width:3px
    style E3 fill:#DC143C,stroke:#8B0000,stroke-width:3px
    style E4 fill:#DC143C,stroke:#8B0000,stroke-width:3px
    style D1 fill:#FFD700,stroke:#000000,stroke-width:3px
    style D2 fill:#FFD700,stroke:#000000,stroke-width:3px
    style D3 fill:#FFD700,stroke:#000000,stroke-width:3px
    style D4 fill:#FFD700,stroke:#000000,stroke-width:3px
```

## Game Flow Examples

```mermaid
sequenceDiagram
    participant Client
    participant Warrior as Warrior Service
    participant Weapon as Weapon Service
    participant Coin as Coin Service
    participant Dragon as Dragon Service
    participant Kafka as Kafka Events
    
    Note over Client,Kafka: Complete Game Flow
    Client->>Warrior: Register & Login
    Warrior-->>Client: JWT Token
    
    Client->>Weapon: Buy Weapon
    Weapon->>Warrior: Validate Token
    Weapon->>Coin: Deduct Coins
    Coin-->>Weapon: Payment Success
    Weapon->>Kafka: Publish Purchase Event
    Weapon-->>Client: Weapon Purchased
    
    Client->>Dragon: Create Dragon (Dark Emperor)
    Dragon-->>Client: Dragon Created
    
    Client->>Dragon: Attack Dragon (Light King)
    Dragon->>Warrior: Get Warrior Info
    Dragon->>Dragon: Calculate Damage
    Dragon->>Kafka: Publish Death Event
    Weapon->>Kafka: Consume Event
    Weapon->>Weapon: Add Loot Weapon
    Dragon-->>Client: Dragon Defeated
```