##Chomba

[![Circle CI](https://circleci.com/gh/dklassen/chamba.svg?style=svg)](https://circleci.com/gh/dklassen/chamba)

// add build kite service  here

### Dependencies

 - Homebrew on OS X
 - Golang 1.6
 - Mysql (will be installed automatically by script/setup)

Follow the sections below to set up your dependencies.

### Initial setup

### Golang

//Describe the golang specific setup

### Mysql
// Describe the brew install process and setup of local environment user for allowing
// chamba to set up the required testing and development databases

## Set up Chamba(OSX)

// Description of how to setup the chamba development environment for running locally
## Development

## Configuration

We encode our secrets using ejson encrypted files. These files are found in `./config`. The config
for a particular environment can be found in the named file `sources.{{environment}}.ejson`.
