# rails-debian
Creates minimal images by compiling the application in a stage and copying only the resulting 
Rails app and its gem into a lightweight base image.


## Configuration
| Field           | Mandatory | Description                                                              |
| --------------- | --------- | ------------------------------------------------------------------------ |
| osRelease       | yes       | Version of Debian.                                                       |
| buildPackages   | yes       | Additional system packages needed during asset compilation               |
| runtimePackages | no        | Additional system packages needed during runtime (e.g. ca-certificates). |

## Project structure
The application's `Gemfile*` has to live inside the root of the project's repository.

## Dependency management
This builder uses Bundle 1.17 for managing dependencies.

## Stages
The builder consists of 4 intermediate stages:
* **base**: with Passenger and the Ruby version specified in the `.ruby-version` file
* **cache-gems-production-only**: extending **cache-packages** with production gems
* **cache-gems-full**: extending **cache-gems-production-only** with all gems
* **cache-packages**: extending **base** with the system packages (`runtimePackages`) and using **cache-gems-full** for asset compilation

The builder has 2 final stages that can be executed:
* **release**: extending **cache-packages** and using **cache-gems-production-only** for the gems and **cache-gems-full** for asset compilation
* **test**: extending **cache-gems-full** and includes the source code. Executes `scripts/test.sh` when started

## Example

```yaml
builderName: rails-debian
builderLocation: ssh://git@github.com:maxlaverse/image-builder
globalSpec:
  osRelease: buster
  buildPackages:
    - libsqlite3-dev
  runtimePackages:
    - libsqlite3-0
    - nodejs
releaseSpec:
  contextInclude:
  - "Gemfile.*"
  - "Rakefile"
  - "app/**"
  - "config/**"
  - "db/**"
  - "lib/**"
  - "public/**"
  - "spec/**"
  - "storage/**"
  - "test/**"
  - "vendor/**"
```

```
$ image-builder build -s test .
[...]
INFO[0004] Build finished! The following images have been pulled or built:
INFO[0004] * generated-rails-signup-thankyou:base-buster-ruby-2.5.1-0894bbea
INFO[0004] * generated-rails-signup-thankyou:cache-gems-production-only-c4e4e61b
INFO[0004] * generated-rails-signup-thankyou:cache-gems-full-1ac36509
INFO[0004] * generated-rails-signup-thankyou:test-d844737e

$ docker run -ti generated-rails-signup-thankyou:test-d844737e
Run options: --seed 59663
# Running:

Finished in 0.002124s, 0.0000 runs/s, 0.0000 assertions/s.
0 runs, 0 assertions, 0 failures, 0 errors, 0 skips
```

```
$ image-builder build .
[...]
INFO[0005] Build finished! The following images have been pulled or built:
INFO[0005] * generated-rails-signup-thankyou:base-buster-ruby-2.5.1-0894bbea
INFO[0005] * generated-rails-signup-thankyou:cache-packages-7728e8c0
INFO[0005] * generated-rails-signup-thankyou:cache-gems-production-only-c4e4e61b
INFO[0005] * generated-rails-signup-thankyou:cache-gems-full-1ac36509
INFO[0005] * generated-rails-signup-thankyou:release-5de2aa4e

$ docker run -eDOMAIN_NAME=domain.com -eSECRET_KEY_BASE=test -eRAILS_ENV=production -p3000:3000 -ti generated-rails-signup-thankyou:release-5de2aa4e
[...]
Starting entrypoint
=============== Phusion Passenger Standalone web server started ===============
PID file: /app/passenger.3000.pid
Log file: /app/passenger.3000.log
Environment: production
Accessible via: http://0.0.0.0:3000/

You can stop Phusion Passenger Standalone by pressing Ctrl-C.
Problems? Check https://www.phusionpassenger.com/library/admin/standalone/troubleshooting/
===============================================================================
[ N 2020-02-18 22:48:50.3478 30/T5 age/Cor/SecurityUpdateChecker.h:519 ]: Security update check: no update found (next check in 24 hours)
I, [2020-02-18T22:48:51.845474 #109]  INFO -- : [83e93095-42e2-4989-a24a-0c45ff2b3344] Started HEAD "/" for 127.0.0.1 at 2020-02-18 22:48:51 +0000
I, [2020-02-18T22:48:51.853413 #109]  INFO -- : [83e93095-42e2-4989-a24a-0c45ff2b3344] Processing by VisitorsController#index as HTML
I, [2020-02-18T22:48:51.857017 #109]  INFO -- : [83e93095-42e2-4989-a24a-0c45ff2b3344]   Rendering visitors/index.html.erb within layouts/application
I, [2020-02-18T22:48:51.863803 #109]  INFO -- : [83e93095-42e2-4989-a24a-0c45ff2b3344]   Rendered visitors/index.html.erb within layouts/application (6.6ms)
I, [2020-02-18T22:48:51.867618 #109]  INFO -- : [83e93095-42e2-4989-a24a-0c45ff2b3344]   Rendered layouts/_navigation_links.html.erb (0.6ms)
I, [2020-02-18T22:48:51.868683 #109]  INFO -- : [83e93095-42e2-4989-a24a-0c45ff2b3344]   Rendered layouts/_nav_links_for_auth.html.erb (0.6ms)
I, [2020-02-18T22:48:51.868754 #109]  INFO -- : [83e93095-42e2-4989-a24a-0c45ff2b3344]   Rendered layouts/_navigation.html.erb (2.7ms)
I, [2020-02-18T22:48:51.869517 #109]  INFO -- : [83e93095-42e2-4989-a24a-0c45ff2b3344]   Rendered layouts/_messages.html.erb (0.4ms)
I, [2020-02-18T22:48:51.869867 #109]  INFO -- : [83e93095-42e2-4989-a24a-0c45ff2b3344] Completed 200 OK in 16ms (Views: 15.2ms)

```