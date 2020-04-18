# python-debian
Creates minimal images by compiling the application in a stage and copying only the resulting 
binary into a lightweight base image.


## Configuration
| Field           | Mandatory | Description                                                              |
| --------------- | --------- | ------------------------------------------------------------------------ |
| osRelease       | yes       | Version of Debian.                                                       |
| pythonVersion   | yes       | Version of Python.                                                           |
| startFile        | yes       | Name of main script to start.                                                     |
| runtimePackages | no        | Additional system packages needed during runtime (e.g. ca-certificates). |
| buildPackages | no        | Additional system packages needed during build (e.g. cmake). |

## Dependency management
This builder uses `pip` for managing dependencies.

## Stages
The builder consists of 3 stages:
* **python-packages** contains the Python packages modules. Only `requirements.txt` is loaded into the build context
* **system-packages** contains the system packages and is used as base for `release`
* **release** contains the source code, Python packages and system packages.

## Example

```yaml
builderName: python-debian
builderLocation: https://github.com:maxlaverse/image-builder
globalSpec:
  osRelease: buster
  pythonVersion: "3.8"
  startFile: "task_import_producer.py"
  buildPackages:
  - cmake
releaseSpec:
  contextInclude:
  - "*.py"
  - "static/*"
```
