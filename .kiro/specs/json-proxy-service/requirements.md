# Requirements Document

## Introduction

This feature implements a configurable JSON REST API proxy service written in Go that accepts requests, transforms them according to configuration, forwards them to target endpoints, and applies response transformations using JSONPath. The service provides a flexible way to proxy and transform API calls with configurable endpoint mappings and JSON response transformations.

## Requirements

### Requirement 1

**User Story:** As a developer, I want to configure target endpoint mappings, so that I can route proxy requests to different backend services.

#### Acceptance Criteria

1. WHEN the service starts THEN it SHALL load configuration from a JSON file that maps unique names to target endpoints (hostname or IP)
2. WHEN a configuration entry is defined THEN it SHALL contain a unique name and a target endpoint URL
3. IF the configuration file is missing or invalid THEN the service SHALL fail to start with a clear error message
4. WHEN the service loads configuration THEN it SHALL validate that all endpoint names are unique

### Requirement 2

**User Story:** As a client, I want to make proxy requests using endpoint names, so that I can access backend services through a unified interface.

#### Acceptance Criteria

1. WHEN a client makes a request to the proxy THEN the first path parameter SHALL correspond to an existing target endpoint name
2. WHEN the endpoint name exists in configuration THEN the proxy SHALL route the request to the corresponding target endpoint
3. WHEN the endpoint name does not exist THEN the proxy SHALL return a 404 error with appropriate message
4. WHEN routing a request THEN the rest of the URL path and query parameters SHALL be passed as-is to the target endpoint

### Requirement 3

**User Story:** As a client, I want to specify request details in JSON body, so that I can control the HTTP method, body content, and response transformation for the target endpoint.

#### Acceptance Criteria

1. WHEN making a proxy request THEN the client SHALL use POST method with JSON body
2. WHEN the JSON body is provided THEN it SHALL contain the HTTP verb for the target endpoint request
3. WHEN the JSON body is provided THEN it SHALL contain the body data for the target endpoint request
4. WHEN the JSON body is provided THEN it SHALL contain a transformation specification for the response
5. IF the JSON body is malformed or missing required fields THEN the proxy SHALL return a 400 error

### Requirement 4

**User Story:** As a client, I want headers to be forwarded to target endpoints, so that authentication and other metadata is preserved.

#### Acceptance Criteria

1. WHEN a proxy request includes headers THEN all headers SHALL be forwarded to the target endpoint
2. WHEN a header is prefixed with "jpx-" THEN it SHALL NOT be forwarded to the target endpoint
3. WHEN forwarding headers THEN the proxy SHALL preserve header names and values exactly

### Requirement 5

**User Story:** As a client, I want to transform JSON responses using JSONPath, so that I can extract and reshape data from target endpoints.

#### Acceptance Criteria

1. WHEN a response is received from the target endpoint THEN the proxy SHALL apply the transformation specified in the request body
2. WHEN the transformation is defined THEN it SHALL use JSONPath expressions to reference fields from the target response
3. WHEN JSONPath expressions are invalid THEN the proxy SHALL return an error with details about the invalid expression
4. WHEN the target response is not valid JSON THEN the proxy SHALL handle the error gracefully
5. WHEN transformation is successful THEN the proxy SHALL return the transformed JSON response

### Requirement 6

**User Story:** As a developer, I want to run the service locally for development, so that I can test and debug the proxy functionality.

#### Acceptance Criteria

1. WHEN running in development mode THEN the service SHALL provide a simple command to start locally
2. WHEN running locally THEN the service SHALL load configuration from a local file
3. WHEN running locally THEN the service SHALL provide clear logging for debugging
4. WHEN configuration changes THEN the developer SHALL be able to restart the service easily

### Requirement 7

**User Story:** As a DevOps engineer, I want to deploy the service as a Docker container, so that I can run it in containerized environments.

#### Acceptance Criteria

1. WHEN building for release THEN the project SHALL produce a Docker image
2. WHEN the Docker image runs THEN it SHALL include all necessary dependencies
3. WHEN running in Docker THEN the service SHALL accept configuration through mounted files or environment variables
4. WHEN the Docker container starts THEN it SHALL expose the proxy service on a configurable port

### Requirement 8

**User Story:** As a developer, I want the configuration system to be abstracted, so that I can switch from file-based to runtime configuration in the future.

#### Acceptance Criteria

1. WHEN implementing configuration loading THEN the code SHALL use an interface or abstraction layer
2. WHEN the configuration interface is defined THEN it SHALL support loading endpoint mappings
3. WHEN implementing file-based configuration THEN it SHALL be one implementation of the configuration interface
4. WHEN switching configuration sources THEN the core proxy logic SHALL remain unchanged
5. WHEN adding new configuration sources THEN it SHALL only require implementing the configuration interface