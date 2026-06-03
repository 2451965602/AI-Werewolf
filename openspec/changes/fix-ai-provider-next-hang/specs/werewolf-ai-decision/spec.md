## MODIFIED Requirements

### Requirement: AI decision generation
The system MUST generate AI decisions without allowing a single slow or malformed provider response to indefinitely block game progression.

#### Scenario: Primary provider times out
- **WHEN** the primary AI provider exceeds its configured timeout during decision generation
- **THEN** the system MUST attempt the same decision through the configured fallback provider

#### Scenario: Primary provider returns unusable content
- **WHEN** the primary AI provider returns empty content or a response that cannot be parsed into the required decision
- **THEN** the system MUST treat that response as a provider failure
- **AND** it MUST attempt the configured fallback provider before giving up

#### Scenario: All configured providers fail
- **WHEN** both the primary and fallback AI providers fail for a decision
- **THEN** the system MUST fall back to the existing non-AI deterministic decision path instead of blocking indefinitely
