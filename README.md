*(Note: This is still a work in progress.
This project is not associated with the
[Bitwarden](https://bitwarden.com/)
project nor 8bit Solutions LLC.)*

## bitwarden-go

[![Build Status](https://travis-ci.org/VictorNine/bitwarden-go.svg?branch=master)](https://travis-ci.org/VictorNine/bitwarden-go)
[![Gitter chat](https://badges.gitter.im//bitwarden-go/Lobby.png)](https://gitter.im/bitwarden-go/Lobby "Gitter chat")

A server compatible with the Bitwarden apps and plugins. The server has a small footprint and could be run locally on your computer, a Raspberry Pi or a small VPS. The data is stored in a local SQLite database.

For more information on the protocol you can read the [documentation](https://github.com/jcs/bitwarden-ruby/blob/master/API.md) provided by [jcs](https://github.com/jcs)

### Usage
#### Fetching the code
Make sure you have the ```go``` package installed.
*Note: package name may vary based on distribution*

You can then run ```go get github.com/VictorNine/bitwarden-go``` to fetch the latest code.

#### Build/Install
Run in your favorite terminal:
```
cd $GOPATH/src/github.com/VictorNine/bitwarden-go/cmd/bitwarden-go
```
followed by
```
go build
```
or
```
go install
```
The former will create a executable named ```bitwarden-go``` in the current directory, and ```go install``` will build and install the executable ```bitwarden-go``` as a system-wide application (located in ```$GOPATH/bin```).
*Note: From here on, this guide assumes you ran ```go install```*

#### Initalizing the Database
*Note: This step only has to be performed once*

Run the following to initalize the database:
```
bitwarden-go -init
```
This will create a database called ```db``` in the directory of the application. Use `-location` to set a different directory for the database.

#### Running
To run [bitwarden-go](https://github.com/VictorNine/bitwarden-go), run the following in the terminal:
```
bitwarden-go
```

#### Usage with Flags
To see all current flags and options with the application, run
```
bitwarden-go -h
```
