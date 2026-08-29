# 008 - Viper Configuration Management with Strongly-Typed Config

**Date**: 2024-01-15  
**Status**: Accepted  
**Decision Area**: Configuration Management  
**TL;DR**: Migrate from ad-hoc environment variable reading to Viper-based configuration management with strongly-typed structs and multi-source loading.

---

## Context

The project previously used scattered `os.Getenv()` calls throughout the codebase for configuration. This approach had several issues:

1. **No Type Safety**: Environment variables are strings; every usage required manual parsing
2. **No Centralization**: Config values were accessed from different parts of the code without a single source of truth
3. **No Documentation**: Discovering all available config options required searching the entire codebase
4. **No Validation**: Invalid configuration values often resulted in cryptic runtime errors
5. **Testing Difficulties**: Mocking individual environment variables scattered across packages was cumbersome

### Existing Configuration Pattern (Before)

```go
// Scattered throughout the codebase
func NewDatabase() (*sql.DB, error) {
    host := os.Getenv("DATABASE_HOST")
    port := os.Getenv("DATABASE_PORT")  // String, needs parsing
    user := os.Getenv("DATABASE_USER")
    password := os.Getenv("DATABASE_PASSWORD")
    name := os.Getenv("DATABASE_NAME")
    
    // Manual validation and error handling everywhere
    if host == "" {
        return nil, errors.New("DATABASE_HOST is required")
    }
    
    portNum, err := strconv.Atoi(port)
    if err != nil {
        portNum = 5432  // Silent default!
    }
    
    // ... connection logic
}

func NewServer() *httptest.Server {
    port := os.Getenv("SERVER_PORT")
    if port == "" {
        port = "8080"  // Another silent default
    }
    // ... server setup
}
```

**Problems**:
- No type checking until runtime
- Defaults scattered across files, hard to discover
- Different validation logic per location
- Hard to test with different configurations

---

## Decision

We adopt **Viper** as the configuration management library with a **strongly-typed config struct** approach.

### Key Components

#### 1. Strongly-Typed Config Structs (`internal/platform/config/config.go`)

```go
type Config struct {
    Database   DatabaseConfig
    Redis      RedisConfig
    JWT        JWTConfig
    Server     ServerConfig
    Logging    LoggingConfig
}

type DatabaseConfig struct {
    Host     string
    Port     int
    User     string
    Password string
    Name     string
}
```

**Benefits**:
- Compile-time type safety for config values
- Single source of truth for all configuration
- IDE autocomplete support
- Easy documentation generation from structs

#### 2. Environment Variable Prefix (`BOXING_`)

All environment variables use the `BOXING_` prefix:

```bash
BOXING_DATABASE_HOST=localhost
BOXING_DATABASE_PORT=5433
BOXING_JWT_SECRET=your-secret-key
```

**Benefits**:
- Namespacing prevents conflicts with other applications
- Clear ownership of environment variables
- Easier shell completion and documentation

#### 3. Multi-Source Loading Hierarchy

Configuration loads from three sources in priority order:

1. **Environment Variables** (highest priority) - `BOXING_*` vars
2. **YAML Config Files** - `config/{environment}.yaml`
3. **Default Values** (lowest priority) - `applyDefaults()` function

```
┌──────────────────────────────────────┐
│ 1. Environment Variables             │  ← Override everything
│    BOXING_DATABASE_PORT=5434         │
└──────────────────────────────────────┘
              ↓ override
┌──────────────────────────────────────┐
│ 2. YAML Config File                  │  ← File-based config
│    config/development.yaml           │
└──────────────────────────────────────┘
              ↓ fallback to
┌──────────────────────────────────────┐
│ 3. applyDefaults()                   │  ← Final safety net
│    Database.Port = 5432              │
└──────────────────────────────────────┘
```

**Benefits**:
- Development: Use YAML files for team-shared settings
- Production: Override sensitive values via environment variables
- Testing: Pure defaults by setting non-existent environment file

#### 4. Environment-Specific Config Files

```yaml
# config/development.yaml
database:
  host: "localhost"
  port: 5433        # Docker-mapped PostgreSQL
  user: "postgres"
  name: "boxing_dev"

server:
  port: 8080

logging:
  level: "debug"
```

```yaml
# config/production.yaml
database:
  host: "${DATABASE_HOST}"  # Must come from env vars
  port: 5432
  
logging:
  level: "error"
```

**Benefits**:
- Team-shared development configuration
- Environment-specific defaults without code changes
- Clear separation of concerns

#### 5. Strict Validation for Required Secrets

Critical values (JWT_SECRET) fail with clear errors:

```go
func validateRequiredFields(cfg *Config) error {
    if cfg.JWT.Secret == "" || cfg.JWT.Secret == "default-jwt-secret-*" {
        return fmt.Errorf("BOXING_JWT_SECRET environment variable is REQUIRED for security")
    }
    // ... other validations
}
```

**Benefits**:
- Fail-fast behavior for missing critical config
- Clear error messages guide developers
- Prevents accidental deployment with default values

#### 6. Test Isolation Pattern

Tests set `BOXING_ENV` to non-existent filenames to test pure defaults:

```go
func TestConfigLoadsDefaults(t *testing.T) {
    _ = os.Setenv("BOXING_ENV", "nonexistent-for-test")
    defer os.Unsetenv("BOXING_ENV")
    
    cfg, err := config.Load()
    assert.NoError(t, err)
    assert.Equal(t, 5432, cfg.Database.Port)  // From applyDefaults()
}
```

**Benefits**:
- Tests verify actual default values from `applyDefaults()`
- No interference from YAML files or environment variables
- Fast, deterministic tests

---

## Alternatives Considered

### Alternative 1: Continue with os.Getenv() + Individual Structs

**Description**: Keep current approach but wrap in local structs per package.

**Pros**:
- Minimal change to existing codebase
- No new dependencies

**Cons**:
- Still scattered across packages
- No central documentation
- Duplicate validation logic likely
- Harder to test with different configs

**Verdict**: Rejected - doesn't solve core problems

### Alternative 2: kingpin/cobra for CLI-based Config

**Description**: Use CLI flag library with config file support.

**Pros**:
- Good CLI integration
- Flag parsing built-in

**Cons**:
- More focused on CLI flags than app config
- Less flexible for nested structures
- Overkill for our use case (minimal CLI)

**Verdict**: Rejected - wrong tool for the job

### Alternative 3: go-viper vs Other Config Libraries

We evaluated these alternatives:

| Library | Pros | Cons | Verdict |
|---------|------|------|---------|
| **Viper** | Mature, feature-rich, env var support, YAML/JSON/TOML | Larger dependency | ✅ Selected |
| **kong** | Great CLI + config combo | Heavy focus on CLI flags | Rejected |
| **pflag + manual parsing** | Standard library only | No structured unmarshaling | Rejected |
| **koanf** | Similar to Viper, lighter weight | Smaller community, less docs | Rejected |

**Viper Selection Rationale**:
- Industry standard for Go configuration
- Excellent environment variable integration
- Strong ecosystem and documentation
- Supports our priority loading pattern natively

### Alternative 4: YAML Files Only (No Env Var Override)

**Description**: Load all config from YAML files only.

**Pros**:
- Simple, explicit configuration
- Easy version control of configs

**Cons**:
- Secrets in YAML = security risk
- No per-environment override flexibility
- Deployments require file changes

**Verdict**: Rejected - security and flexibility concerns

### Alternative 5: Full Environment Variable Only (No Files)

**Description**: Use only environment variables, no YAML files.

**Pros**:
- 12-factor app compliant
- No file management
- Easy CI/CD integration

**Cons**:
- Hard to discover all available config options
- No team-shared development defaults
- Every developer must set all env vars locally

**Verdict**: Rejected - usability issues for local development

---

## Consequences

### Positive

1. **Type Safety**: Compile-time checking of config structure prevents runtime surprises
2. **Centralized Documentation**: `Config` struct serves as single source of truth; can auto-generate docs
3. **Clear Defaults**: `applyDefaults()` explicitly shows fallback values in one location
4. **Security**: Secrets come only from environment variables, never committed files
5. **Testing**: Easy to test with different configurations via `BOXING_ENV` and env var mocks
6. **Discoverability**: New developers can see all config options in one file

### Negative

1. **New Dependency**: Adding `github.com/spf13/viper` (~2MB) to the project
2. **Learning Curve**: Team needs to understand loading hierarchy (YAML → Env → Defaults)
3. **Migration Effort**: Existing code using `os.Getenv()` must be refactored
4. **Debugging Complexity**: Determining which source set a value requires understanding all three layers

### Mitigations

1. **Dependency Size**: Acceptable trade-off for the functionality provided; Viper is widely trusted
2. **Documentation**: Comprehensive guide in `docs/configuration.md` explains loading order
3. **Migration**: Phased approach - new code uses config, gradually migrate old code
4. **Debugging**: Add debug logging showing final config values at startup

---

## Implementation Details

### File Structure

```
internal/platform/config/
├── config.go          # Strongly-typed structs and validation
├── viper.go           # Viper setup and environment binding
├── config_test.go     # Unit tests for config loading
└── default.go         # Default value definitions (if separated)

config/
├── development.yaml   # Development environment defaults
├── test.yaml          # Test environment configuration  
└── production.yaml    # Production environment template

.env.example           # Environment variable template (committed)
.env.local            # Local secrets (gitignored)
```

### Config Loading Sequence

```go
func Load() (*Config, error) {
    // 1. Create Viper instance with BOXING_ prefix
    v := newViper()
    
    // 2. Read YAML file (optional - based on BOXING_ENV)
    // development.yaml if BOXING_ENV=development
    
    // 3. Bind environment variables
    readEnvConfig(v)
    
    // 4. Unmarshal into struct
    var cfg Config
    v.Unmarshal(&cfg)
    
    // 5. Apply defaults for any empty values
    applyDefaults(&cfg)
    
    // 6. Validate required fields
    if err := validateRequiredFields(&cfg); err != nil {
        return nil, err
    }
    
    return &cfg, nil
}
```

### Environment Variable Naming Convention

```
YAML Key              → Environment Variable
─────────────────────────────────────────────
database.host         → BOXING_DATABASE_HOST
database.port         → BOXING_DATABASE_PORT
jwt.secret            → BOXING_JWT_SECRET
logging.level         → BOXING_LOGGING_LEVEL
```

Rule: `BOXING_` + uppercase key with dots replaced by underscores.

---

## Testing Strategy

### Unit Tests for Default Values

```go
func TestDefaultConfiguration(t *testing.T) {
    // Prevent YAML loading
    _ = os.Setenv("BOXING_ENV", "nonexistent")
    defer os.Unsetenv("BOXING_ENV")
    
    cfg, err := Load()
    assert.NoError(t, err)
    
    // Verify defaults from applyDefaults()
    assert.Equal(t, 5432, cfg.Database.Port)
    assert.Equal(t, "info", cfg.Logging.Level)
}
```

### Integration Tests with Real Config Files

```go
func TestDevelopmentConfiguration(t *testing.T) {
    _ = os.Setenv("BOXING_ENV", "development")
    defer os.Unsetenv("BOXING_ENV")
    
    cfg, err := Load()
    assert.NoError(t, err)
    
    // Verify YAML values are loaded
    assert.Equal(t, 5433, cfg.Database.Port)  // From development.yaml
}
```

### Environment Variable Override Tests

```go
func TestEnvVarOverridesYAML(t *testing.T) {
    _ = os.Setenv("BOXING_ENV", "development")
    _ = os.Setenv("BOXING_DATABASE_PORT", "9999")
    defer func() {
        os.Unsetenv("BOXING_ENV")
        os.Unsetenv("BOXING_DATABASE_PORT")
    }()
    
    cfg, err := Load()
    assert.NoError(t, err)
    
    // Env var should override YAML
    assert.Equal(t, 9999, cfg.Database.Port)
}
```

---

## Migration Checklist

- [x] Create `internal/platform/config` package with Viper setup
- [x] Define strongly-typed config structs in `config.go`
- [x] Implement `applyDefaults()` for fallback values
- [x] Add validation in `validateRequiredFields()`
- [x] Create `config/development.yaml` with team defaults
- [x] Write comprehensive tests covering all loading scenarios
- [x] Update `.env.example` with documented template
- [x] Update `.gitignore` to exclude `.env*` but keep examples
- [ ] Migrate existing `os.Getenv()` calls to use central config
- [ ] Add documentation in `docs/configuration.md`
- [ ] Update README.md with configuration quick start

---

## Related Issues and Decisions

- **MAT-62**: Refactor Configuration Management to Viper with Strongly-Typed Config
- **MAT-46**: Setup isolated test database for integration tests (enabled by this config system)
- **[001-technology-stack.md](001-technology-stack.md)**: Initial technology choices
- **[006-architecture-patterns.md](006-architecture-patterns.md)**: Relates to platform abstraction layer

---

## References

- [Viper Documentation](https://github.com/spf13/viper)
- [Twelve-Factor App Config](https://12factor.net/config)
- `internal/platform/config/` package implementation
