# Installation

Go package management is a mess right now. gopkg.in was a band-aid and 
worked well if dependencies were on the master branch or followed gopkg.in
convention. The problem is and has always been how do you lock down to 
a specific commit or tag? That's where glide shines.

## Pre-requisites

Install [glide](https://github.com/Masterminds/glide)

## Installing to $GOPATH

```sh
cd $GOPATH
git clone -b v1 https://github.com/mgutz/dat gopkg.in/mgutz/dat.v1
cd gopkg.in/mgutz/dat.v1
glide install
```

## Existing Glide Project

Inside the project

```sh
glide get gopkg.in/mgutz/dat.v1
```
