# Distributed system for dependent GitHub CI/CD gating

## About

The code of the server side of the dependent gating system.

Related repositories:
- [Repository with test data](https://github.com/polupanovaanna/github_actions_test_project)
- [GitHub action for data translation](https://github.com/polupanovaanna/parse_repo_docker_action)

## How to run

To start the server run the following command: `go run main.go`

To run test case: `go run client/test_client.go`

Your server will be running on the localhost:8080
