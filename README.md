[![Build Status](https://travis-ci.org/nnev/frank.svg?branch=robust)](https://travis-ci.org/nnev/frank)

### Intro

While it’s possible to install this go package using the common way it likely won’t work: Some of configuration is hard coded as constants, so you most likely simply want to check out the repository, modify and run `go install` to get your binary of choice. 

frank connects directly to [RobustIRC networks](https://robustirc.net/) using the [offical bridge implementation](https://github.com/robustirc/bridge) to translate between IRC and RobustIRC formats.

### Installation

```
go get github.com/nnev/frank
echo "Modify away!"
```

### Attribution

The project is ISC-licensed, but all other software used remains under their respective license.

- Go, see http://golang.org/LICENSE
