Development environment for the Morpheo project
===============================================

This repository holds a docker-compose environment for the Morpheo project. It
embeds all the different projects' code and libs as Git submodules. You can hack
on the project's code directly in Git submodules. For developers with direct
push right to this repository, please be sure to read and understand 100% of the
[Git documentation about submodules](https://git-scm.com/book/en/v2/Git-Tools-Submodules) for God's sake !

It also contains a Makefile that detects changes in Git submodules and
automatically rebuilds what needs to be rebuilt (and that only) and updates the
dev. environment. Interpreted code (Python, frontend code) should be mounted as
a volume directly in the appropriate container, therefore you should have to
rebuild only when you changes dependencies.

Getting started
---------------

Make sure `make`, `docker` and `docker-compose` are installed on your machine
before going any further.

#### TODO: detail how to do this
Spawning the dev. env. after rebuilding all components that need to
Building a Golang binary
Updating dependencies (glide/pip)
Running tests for a given project
Building Docker images

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
