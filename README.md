# image-builder

Quickly build a Container Image out of an application's source code.

`image-builder` is a command line tool that helps building Container images without having to write .
On the one side there is a YAML file named Build Configuration, specific to an application, that defines some
settings for the resulting Container Image. On the other side, there is a Builder Definition which is a set of
templates used to generate  and build Container Images. A single Builder can use multiple 
to build intermediary images and optimize caching. The final image(s) is put together by using a standard
multi-stage build with the [`COPY --from` directive][dockerfile-copy].

Only [Docker][docker-website] and [Podman][podman-website] are supported as Container Engine.

----

**Disclaimer: This project is experimental.**

----

## Table of contents
* [Prerequisites](#prerequisites)
* [Usage](#usage)
* [Concepts](#concepts)
  * [Build Configuration](#build-configuration)
  * [Builder Definition](#builder-definition)
* [Cache invalidation](#cache-invalidation)
* [Prebuilding stages](#prebuilding-stages)
  * [Builder Cache](#builder-cache)
  * [Prepare Stages](#prepare-stages)
* [Anatomy of a build](#anatomy-of-a-build)

----

## Prerequisites
* The `image-builder` binary
* An application with a [Build Configuration](#build-configuration)
* A [Builder Definition](#builder-definition) for the type of application to be built (e.g [Go][builder-goapp], [Rails][builder-railsapp])
* A Container Engine (Docker and Podman are supported)

## Usage
```
$ git clone git@github.com/maxlaverse/example-of-application.git
$ cd example-of-application
$ cat <<EOF > build.yaml
builder:
  name: go-debian
  location: ssh://git@github.com:maxlaverse/image-builder-collection.git[#branch:[subfolder]]
EOF

$ image-build build .
[...]
```

## Concepts

### Build Configuration
The Build Configuration is a YAML file, usually specific to an application and commited in its repository. It contains
the required settings to build a Container Image out of the source code of an application. There are two mandatory
information:
* `builder.name`: the name of the Builder which is like the type of the application (e.g: Go, Ruby, Scala)
* `builder.location`: the location of the Builders (e.g filesystem, git repository)

**Example:**
```
builder:
  name: go-debian
  location: ssh://git@github.com/maxlaverse/image-builder.git#master:builders

  # [optional] Image registry to lookup for commonly used cache images
  cache: docker.io/maxlaverse

# Additional settings for the Dockerfile generation
imageSpec:
  osRelease: bionic
  passengerVersion: 6.0.22
  runtimePackages:
  - ca-certificates
  - gzip
```

### Builder Definition
A Builder is a set of stages that are required to transform an application of a given type (e.g Go, Ruby, NodeJS) into a container image.

#### Folder structure
A Builder Definition is a folder that holds one or multiple subfolders. Each of those subfolders represents a stage and
contains a Dockerfile as well as additional files to be included in the corresponding Container Images.

**Example:**
```
.
└── goapp
    ├── cache-modules
    |   └── Dockerfile            # Image with all the Go module downloaded
    ├── cache-system-packages
    |   └── Dockerfile            # Image with the system packages pre-installed
    └── release
        ├── entrypoint.sh
        └── Dockerfile            # Multi-stage build depending on the other stages
```

#### Stages
Each Buidler has at least one stage named *release*. The main advantage of usage multiple stages it to split an
application into multiple parts that can each be cached individually to make consecutive builds faster. One very
common stage is a *dependency stage* that contains all dependencies an application requires (e.g Gem, Go module) during
compilation. 

The stages Dockerfiles declare how they depend on each other in order for `image-builder` to build them in the right
order. Before `image-builder` tries to build a Container Image for a given stage, it computes a Content Hash which is a
checksum of the data in the Build Context, including the content of the generated Dockerfile. It then verifies if an
image is already available with the same Content Hash and can be pulled. If this is not the case, the stage image is
built.

At the end of the execution, each stage that was built is pushed into an image registry with a tag matching its
Content Hash.

#### Builder Templating
The Dockerfiles of a Builder use Go templating features. This allows to dynamically generate part of the Dockerfile
based on the source code, and the settings specified in the application's Build Configuration.

##### Helpers
A few functions are available on top of what the Go template language already provides.

| Name                                     | Description                                             | Example                                    |
|------------------------------------------|---------------------------------------------------------|--------------------------------------------|
| `BuilderStage(stageName)`                | Return the generated image name for a given stage       | `FROM {{BuilderStage "cache"}} AS builder` |
| `ExternalImage(imageName)`               | Return the SHA fingerprint of an image.                 | `FROM {{ExternalImage "debian:buster"}} AS baseLayer` |
| `GitCommitShort()`                       | Return the current Git commit                           | `RUN echo "{{GitCommitShort}}" > /app/REVISION` |
| `HasFile(filepath)`                      | Check if a file is present in the **local** context     |                                            |
| `Parameter(parameterName)`               | Return a given field of the `spec`                      | `RUN apt-get update && apt-get install -y {{range $val := (Parameter "runtimePackages")}}{{$val}} {{end}}` |
| `MandatoryParameter(stageName)`          | Return a given field of the `spec` or failed            | `ENTRYPOINT ["/bin/{{MandatoryParameter "binary"}}"]` |
| `File(filepath)`                         | Return the content of a file from the **local** context |                                            |
| `ImageAgeGeneration(imageName, duration)`| Returns the image age divided by the specific duration  |                                            |

Note that `BuilderStage` and `ExternalImage` should always be prefered over hard-coding an image name as they
play an important role in dependency resolution and content cache invalidation. `BuilderStage` ensures stages
are build in the right order, and by replacing an image with its digest, `ExternalImage` makes sure a stage is rebuilt
if the parent image changes.

##### Directives
A `Dockerfile` can also include additional directives written as comments. They help tunning the build process and can
play a role in cache invalidation. They have the form of `# Key` or `# Key Value`.

| Name                    | Description                                                                      |
|-------------------------|----------------------------------------------------------------------------------|
| `DockerIgnore`          | Adds an item to the .dockerignore file that is generated during the build        |
| `UseBuilderContext`     | Use the Builder's folder as build context instead of the application's folder. Required if the stage is embedding files from the Builder's folder.|
| `FriendlyTag`           | Appends a friendly information to the tag (e.g os release, package version)      |
| `TagAlias`              | Push the resulting image with extra tag (e.g: v2, v2.6, v2.6.5)                  |
| `ContentHashIgnoreLine` | Tells the Content Hashing algorithm to ignore the next line. Useful if the next line is dynamic (e.g `GitCommitShort()`) |

## Cache invalidation
The Content Hashing alrorithm is at the center of the image cache management. What ever changes the value of the
Content Hash leads to the stage image to be rebuilt.

Depending on the Build Configuration and Builder Definition, the following condition may change the Content Hash:
* the content of the generated `Dockerfile` is changed, e.g:
  * if `FROM` uses `ExternalImage()` and the corresponding image digest changed
  * if `FROM` uses `BuilderStage()` and the Content Hash of the other stage changed
  * when the Dockerfile template itself changed (update of the Builder definition)
  * when a value used to render the Dockerfile changed (e.g version of a system package to install)
* the content of the Build Context changed (if not ignored with `.dockerignore`)

As always with Container Image build, some layers may result in different images depending when then run.
This is the case when `apt-get update` is executed during the build, or any `wget` or command line interacting with
resources external to the build process. To avoid unpleasant surprises, avoid such layer when possible.
In case of emergency, to force all users to re-run such a command you can invalidate all the caches by changing
anything in a Builder's definition.

## Prebuilding stages

### Builder Cache
Before a stage is built, `image-builder` look into the application's image registry if an image is already available.
Users have the possibility to define an additional registry URL in their Build Definition to lookup for cached images.
This allows to build some specific stages and have them shared with everyone, instead of having each user caching its
own version of the same stage.

Those images are sometimes refered as *prebuild* images. Good candidates are stage that don't include any source code
but only install system packages (e.g an Ubuntu image with Go). The command `image-builder prebuild` helps generating
the prebuilt images.

### Prepare stages
Depending on the Builder and the type of test, it makes sense to prebuild some of the stages as a first step of a
CI/CD pipeline. This is especially relevant if a stage is not used to produce a release image, but to mount the
source code and run some tests inside a container that already has all dependencies installed. To parallelize those
tests, the test stage image needs to be available already.

This can easily be achieved by running `image-builder build -s cache -s test`

## Anatomy of a build
Given that you have properly installed `image-builder`, that the Docker daemon or Podman is available
and that your application has a Build configuration, you should be able to execute:
```
$ image-builder build .
```

First `image-builder` ensures that you have the latest version of the Builder definitions. If the location
is a Git repository, `image-builder` will either clone it or pull it.

It then verifies that the content of the Builder is valid and renders the `Dockerfile` for each available stage in
order to build a dependency graph. For each stage, it computes a Content Hash of the current context and then tries to
find an image with the expected tag on the Builder image registry first (if `builderCache` has been specific in the Build
Configuration). If it can't be found, a second try is done on the application's image registry. Ultimately, the image
for the stage is either pulled or built. When a stage needs to be built, `image-builder` pushes the resulting image to
the application's image registry.

## TODOs

### CLI
* Verify the push decision is alright
* Remove all the TODOs
* Mention origin of `import "github.com/docker/docker/pkg/fileutils"`
* Add a configuration files to allow changing the defaults
* Implement `image-builder config *`
* Allow to overlay builders
* Command to prune cache for an app, to prune baseLayers, manually
* Allow to use wildcards when specifying stages to build
* Explain cache invalidation, apt-get and how ImageAgeGeneration might help (and choose a better name for it)
* Specify default image in build.yaml ?
* Add tests
* Flag in `prebuilt.yaml` to delete an image from registry  ?

### Builders
* Add LABELs in pre-build images ?
* Expiration labels for certain stages ? As variable of preBuilt ?

[dockerfile-copy]: https://docs.docker.com/engine/reference/builder/#copy
[docker-website]: https://docs.docker.com/
[podman-website]: https://podman.io/
[builder-goapp]: builders/go-debian/README.md
[builder-railsapp]: builders/rails-debian/README.md
