*(Note: This is still a work in progress.
This project is not associated with the
[Bitwarden](https://bitwarden.com/)
project nor 8bit Solutions LLC.)*

## bitwarden-go

[![Build Status](https://travis-ci.org/VictorNine/bitwarden-go.svg?branch=master)](https://travis-ci.org/VictorNine/bitwarden-go)

A server compatible with the Bitwarden apps and plugins. The server has a small footprint and could be run locally on your computer, a Raspberry Pi or a small VPS. The data is stored in a local SQLite database.

For more information on the protocol you can read the documentation provided by jcs https://github.com/jcs/bitwarden-ruby/blob/master/API.md

### Usage
Run ./bitwarden-go -init to initialize the database. This only needs to be done once. The server should now be running on port 8000.
