[![Build Status](https://travis-ci.org/breunigs/frank.svg?branch=master)](https://travis-ci.org/breunigs/frank)

### Intro

While it’s possible to install this go package using the common way it likely won’t work: The configuration is hard coded as constants, so you most likely simply want to check out the repository, modify and run `go install` to get your binary of choice.

### Installation

```
apt-get install liburi-find-perl
go get github.com/breunigs/frank
echo "Modify away!"
```

Please note that you need the master branch of the goirc library for frank to compile, i.e. `cd gocode/src/github.com/fluffle/goirc && git checkout master`.

### Attribution

The project is ISC-licensed, but all other software used remains under their respective license.

- Go, see http://golang.org/LICENSE
- `goirc` © Alex Bramley; same license as Go
