## ADDED Requirements

### Requirement: Runtime configuration source priority

The application SHALL resolve runtime configuration using the priority order: environment variables, then configuration file values, then built-in defaults.

#### Scenario: Environment variable overrides configuration file

- **GIVEN** a configuration file defines `storage.state_path` as `data/from-file.json`
- **AND** `WEREWOLF_STORAGE_STATE_PATH` is set to `data/from-env.json`
- **WHEN** the application loads configuration
- **THEN** the resolved storage state path is `data/from-env.json`

#### Scenario: Configuration file overrides default value

- **GIVEN** no storage path environment variable is set
- **AND** a configuration file defines `storage.state_path` as `data/from-file.json`
- **WHEN** the application loads configuration
- **THEN** the resolved storage state path is `data/from-file.json`

#### Scenario: Defaults apply when no external configuration exists

- **GIVEN** no configuration file exists at the default path
- **AND** no related environment variables are set
- **WHEN** the application loads configuration
- **THEN** the resolved server address and storage path use built-in defaults

### Requirement: Configurable HTTP server address

The application SHALL use the resolved server address when constructing the HTTP server.

#### Scenario: Custom server address is configured

- **GIVEN** `WEREWOLF_SERVER_ADDR` is set to `:9090`
- **WHEN** the application starts
- **THEN** the HTTP server is configured to listen on `:9090`

### Requirement: Configurable persistent state path

The application SHALL use the resolved storage state path for JSON state persistence.

#### Scenario: Custom state path is configured

- **GIVEN** `WEREWOLF_STORAGE_STATE_PATH` is set to `tmp/state.json`
- **WHEN** the application creates its JSON store
- **THEN** the store uses `tmp/state.json` as the persistence path

### Requirement: AI secret configuration

The application SHALL allow the AI API key to be provided by direct environment variable, configuration file value, or an environment variable name configured in `ai.api_key_env`.

#### Scenario: API key direct environment variable overrides configuration file

- **GIVEN** the configuration file defines `ai.api_key` as `file-secret`
- **AND** `WEREWOLF_AI_API_KEY` is set to `env-secret`
- **WHEN** the application loads AI configuration
- **THEN** the resolved AI API key is `env-secret`

#### Scenario: API key can be stored in configuration file

- **GIVEN** no direct AI API key environment variable is set
- **AND** the configuration file defines `ai.api_key` as `file-secret`
- **WHEN** the application loads AI configuration
- **THEN** the resolved AI API key is `file-secret`

#### Scenario: API key environment variable name is configured as fallback

- **GIVEN** the configuration file defines `ai.api_key_env` as `OPENAI_API_KEY`
- **AND** `OPENAI_API_KEY` is set in the process environment
- **AND** no direct AI API key value is configured
- **WHEN** the application loads AI configuration
- **THEN** the resolved AI API key is read from `OPENAI_API_KEY`
