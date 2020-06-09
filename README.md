# RedisSeer

![GitHub go.mod Go version (subfolder of monorepo)](https://img.shields.io/github/go-mod/go-version/bvisonl/redis-seer) ![GitHub tag (latest by date)](https://img.shields.io/github/v/tag/bvisonl/redis-seer) [![Build Status](https://travis-ci.com/bvisonl/redis-seer.svg?token=HCrJv9pAbBM8v4eissjP&branch=master)](https://travis-ci.com/bvisonl/redis-seer) ![GitHub stars](https://img.shields.io/github/stars/bvisonl/redis-seer)


> :warning: This is currently just a proof of concept and it came up from an exercise I wanted to do because I was not able to find a simple Redis proxy/HA that worked for standalone environments that supported both balancing and failover (like how AWS ElastiCache works).

## Requirements

* `Go` if you want to build the project from source (otherwise download one of the releases)


## Installation
```go
# Download from the repository
git clone https://github.com/bvison/redis-seer.git
cd redis-seer
go build
```

You should see a `redis-seer` binary inside the folder

## Configuration

`redis-seer` feeds its configuration from a YAML file  that. This is a full sample configuration:

```yaml
---
debug: true
port: 5100
mode: "pick-one" # round-robin, random, pick one, load
readSlaveOnly: true
monitorInterval: 1
db: 0
servers:
  redis1:
    enabled: true
    alias: redis1
    host: "192.168.237.72"
    port: 7000
  redis2:
    enabled: true
    alias: redis2
    db: 0
    host: "192.168.237.72"
    port: 7001

```

Some parameters that may need explanation:

* **port:** Port on which `RedisSeer` will listen for requests.

* **readSlaveOnly:** This will avoid sending commands such as `GET` to the current master.

* **monitorInterval:** The interval on which the servers will be contacted for things like: checking the master, pinging the slaves, and also reconnecting when a connection drops.

* **db:** Database to use when connecting to the redis servers. As of right now if multiple databases are needed additional instances of `RedisSeer` with different configuration files should be deployed.

* **mode:** The selection process to be used when picking a server to send the command to.
    * **round-robin**: Typical round-robin method, the servers will be selected in a sequential way.
    * **random**: A random server will be picked from the list.
    * **pick-one** *(default)*: This is the default, a random server will be picked and it will be used for future commands until that connection drops then the process will be repeated.
    * **load**: The proxy will keep note on the load sent to the server and will pick the least targeted server.


## Usage

**On Linux**:
```bash
$ ./redis-seer -c config.yml
```
**On Windows**:
```bash
$ redis-seer.exe -c config.yml
```

## Example

Inside the `tests/redis` folder there is a `docker-compose.yml` file that boots up 3 redis instances to be used with the tests and if you want to you can test `redis-seer` with them.

**Run Redis servers**
```bash
cd tests/redis
docker-compose up -d

# Run docker ps to confirm that the servers are up and running
```

Also, inside `tests/` there is a `config.yml` with the sample configuration to be used with this `docker-compose.yml`. If needed, modify it at will or point the parameter to the correct `yml` file.


**Run RedisSeer**
```bash
./redis-seer -c tests/config.yml
```

## TODO

* Docker deployment
* Implement selection method
* Add Master monitoring process
* Add method to perform failover on master failure
* Add config file as an argument
* Organize code structure
* Make use of channel
* Improve error handling
* Handle connection changes appropriately
* Add tests
* Run benchmarks
* Ansible deployment

## Contribution
