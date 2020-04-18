# go-debian
Creates minimal images by compiling the application in a stage and copying only the resulting 
binary into a lightweight base image.


## Configuration
| Field           | Mandatory | Description                                                              |
| --------------- | --------- | ------------------------------------------------------------------------ |
| osRelease       | yes       | Version of Debian.                                                       |
| goVersion       | yes       | Version of Go.                                                           |
| goBinary        | yes       | Name of main binary.                                                     |
| runtimePackages | no        | Additional system packages needed during runtime (e.g. ca-certificates). |
| goBuildCmd      | no        | Alternative command to build the binary (default: `go build -mod=readonly`). |
| goGetModules    | no        | Additional modules to download in the module stage. |

## Project structure
The application's main Go file has to live inside the root of the project's repository.

## Dependency management
This builder uses `go mod` for managing dependencies.

## Stages
The builder consists only in two stages:
* **cache** contains the Go modules. The Content Hash only takes `go.mod` and `go.sum` into account
* **release** contains the compiled Go application

## Example

```yaml
builderName: go-debian
builderLocation: ssh://git@github.com:maxlaverse/image-builder
globalSpec:
  osRelease: buster
  goVersion: "1.13"
  binary: ovh-exchange-backup
  runtimePackages:
  - ca-certificates
releaseSpec:
  contextInclude:
  - "pkg/**/*.go"
```
