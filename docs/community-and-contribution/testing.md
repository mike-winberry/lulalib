# Testing

## Execution

There are multiple layers of tests that can be run in isolation or collectively as a whole. 

- `make test` will execute the full suite of tests
  - This requires an environment with `kind` available
- `make test-e2e` will execute the end-to-end tests for CLI and Kubernetes testing
  - This requires an environment with `kind` available
- `make test-cmd` tests the CLI tests
  - Does not require additional infrastructure
- `make test-unit` runs the unit tests
  - Does not require additional infrastructure

## Test Data

Testing artifacts are stored in relation to the tests being run or centralized for access across multiple tests. All `.golden` files are generated with `go test <path> -update` and should not be modified manually. 
