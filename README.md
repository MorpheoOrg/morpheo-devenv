# Development Environment for the Morpheo project

This repository holds a docker-compose environment for the Morpheo project.

It also contains a Makefile that detects changes in those repositories and
automatically rebuilds what need to be rebuilt (and that only) and updates the
dev. environment.

## Table of Content
- [Setup](#setup)
    - [Requirements](#requirements)
    - [Set the directory tree](#set-the-directory-tree)
    - [Start the Development Environment](#start-the-development-environment)
- [Usage](#usage)
- [Tests](#tests)

## Setup
#### Requirements
* [Go](https://golang.org/doc/install) version >= 1.8
* [dep](https://github.com/golang/dep) the official Go dependency management tool. You can install it by running `go get github.com/golang/dep/cmd/dep`
* [Docker](https://docs.docker.com/engine/installation/) and [Docker Compose](https://docs.docker.com/compose/install/)
* [GNU Make](https://www.gnu.org/software/make/)

#### Set the directory tree
To build and launch the Morpheo services, the development environment searches for their respective git repositories **in the parent directory**. Consequently, the directory architecture should be like this:
```
$GOPATH/src/github.com/MorpheoOrg
                                |___morpheo-devenv
                                |___morpheo-compute
                                |___morpheo-storage
                                |___morpheo-go-packages
                                |___morpheo-orchestrator-bootstrap

```


To set everything up, you can run the following in your `$GOPATH/src/github.com` directory:
```
mkdir MorpheoOrg &&
cd MorpheoOrg &&
git clone https://github.com/MorpheoOrg/morpheo-devenv.git &&
git clone https://github.com/MorpheoOrg/morpheo-compute.git &&
git clone https://github.com/MorpheoOrg/morpheo-storage.git &&
git clone https://github.com/MorpheoOrg/morpheo-go-packages.git &&
git clone https://github.com/MorpheoOrg/morpheo-orchestrator-bootstrap.git
```


#### Start the Development Environment
Once the directory architecture is in place, you have to follow the instructions
of `morpheo-orchestrator-bootstrap` to set up a fabric network.

Once the fabric network is set, you can launch the network by running in the `morpheo-devenv` repository:
```
make network
```

Then you can launch the Compute and Storage Morpheo services by running:
```
make up
```

The first time may last quite a while, as all the libraries and the docker images need to be pulled.

Once `make up` has run, you can check with your favourite tool (such as `ctop`) that the containers have been properly launched. To see Morpheo in action, run the [integration tests](#tests).

Note that the exposed ports for the services can be changed in the Makefile, the default one being:
* Storage: 8081
* Compute: 8082

## Usage
GNU Make is used to interact with the devenv:

##### Fabric network
* `make network`: **start the network**, by running a `./byfn.sh -m up -i` in `morpheo-orchestrator-bootstrap`
* `make network-down`: **clean the network**, by running a `./byfn.sh -m down`

##### Compute and Storage
* `make up`: **start compute and storage**, updating the vendor, building the binaries and running a `docker-compose up`
* `make stop`: **stop all the containers**, by running `docker-compose stop`
* `make logs`: **show the logs of the main containers**, by running `docker-compose logs`
* `make down`: **delete all the containers**, by running `docker-compose down`
* `make clean`: **delete all the containers and the data**, including storage files, postgres and mongo data
* `make tests`: **run the integration tests**


Note that `make up` does the following when needed:

1. Update *compute vendor* and *storage vendor* folder with `dep ensure`.
2. Replace the folder `morpheo-go-packages` in *compute vendor* and *storage vendor* by your local folder in the parent directory `MorpheoOrg/morpheo-go-packages`. This step is crucial for development, as `dep` fetches the latest github release of morpheo-go-packages and **not** your local repository. Consequently, if you are working on go-packages and you want to tests the change you have made, this replacement is necessary.
3. Build the Compute and Storage Go binaries
4. Run `docker-compose up` with variables set in the Makefile to build the docker images and launch the Morpheo services in containers


## Tests
The Devenv provide a script tests/integration.go that tests that the whole plateform works.

##### Setup
To perform the tests, you have to download the associated data from Google Cloud Storage and put it in the right folder:
```
cd morpheo-devenv/data   # default path for the fixtures data
wget https://storage.googleapis.com/morpheo-storage/fixtures/fixtures.tar.gz
tar -zxvf fixtures.tar.gz && rm fixtures.tar.gz
```

##### Usage
```
make tests
```


License
-------

All this code is open source and licensed under the CeCILL license - which is an
exact transcription of the GNU GPL license that also is compatible with french
intellectual property law. Please find the attached licence in English [here](./LICENSE) or
[in French](./LICENCE).

Note that this license explicitely forbids redistributing this code (or any
fork) under another licence.

Maintainers
-----------
* Étienne Lafarge <etienne_a t_rythm.co>
* Max-Pol Le Brun <maxpol_a t_morpheo.co>