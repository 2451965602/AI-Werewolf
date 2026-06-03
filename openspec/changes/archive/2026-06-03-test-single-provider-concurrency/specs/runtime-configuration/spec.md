## MODIFIED Requirements

### Requirement: AI provider runtime selection

The application SHALL construct the AI provider from resolved runtime configuration instead of hardcoding a fallback provider.

#### Scenario: Configured provider is selected at startup

- **GIVEN** runtime configuration resolves `ai.provider` to a supported provider name
- **WHEN** the application starts
- **THEN** the service uses the configured provider implementation for AI calls

#### Scenario: Unknown provider fails fast

- **GIVEN** runtime configuration resolves `ai.provider` to an unsupported provider name
- **WHEN** the application starts
- **THEN** startup fails with a clear configuration error

### Requirement: AI concurrency configuration

The application SHALL support `ai.concurrency` as the maximum number of concurrent calls allowed for the configured AI provider instance.

#### Scenario: Concurrency one serializes provider calls

- **GIVEN** runtime configuration resolves `ai.concurrency` to `1`
- **WHEN** multiple AI calls are made concurrently through the same provider instance
- **THEN** the calls execute one at a time in serial order

#### Scenario: Invalid concurrency fails fast

- **GIVEN** runtime configuration resolves `ai.concurrency` to `0` or a negative number
- **WHEN** the application starts
- **THEN** startup fails with a clear configuration error
