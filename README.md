# aws-rekey
[![Build Status](https://travis-ci.org/vaijab/aws-rekey.svg?branch=master)](https://travis-ci.org/vaijab/aws-rekey)

Re-key your static AWS API access keys.

A lot of the times, people will just have their AWS access keys stored in the
shared credentials file permanently. It can be very tedious to manually
generate new keys, update the credentials file and delete the old keys, so in
the end we rarely end up doing that.

aws-rekey allows one to roll the access keys and save them to the shared
credentials file i.e. `~/.aws/credentials`.


## Installation

Grab the latest arch-specific binary from [releases
page](https://github.com/vaijab/aws-rekey/releases/latest) and save it as
`${HOME}/bin/aws-rekey`.


## Usage

By default aws-rekey will attempt to find your credentials file, but you can
also specify its location, using `--credentials-file` parameter.

AWS access keys can be rolled for multiple profiles, by specifying comma
separated profile names, e.g. `aws-rekey --profiles my-test,production,development`


## Contribution

Please do!


## Author

- Vaidas Jablonskis <jablonskis@gmail.com>

