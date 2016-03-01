# GooGet 
[![Build Status](https://travis-ci.org/google/googet.svg?branch=master)](https://travis-ci.org/google/googet)

GooGet (Googet's Obviously Only a Goofy Experimental Title) is a modular
package repository solution primarily designed for Windows. 

This is not an official Google product.

## Build
Run build.cmd/build.sh to build GooGet for Windows. To package googet run 

```
go run goopack/goopack.go googet.goospec
```

This will result in googet.x86_64.VERSION.goo which can be installed on a 
machine with the `googet install` command (assuming googet is already 
installed).

To install on a fresh machine copy both googet.exe and the googet package
over and run:

```
googet -root 'c:/ProgramData/GooGet' install googet googet.x86_64.VERSION.goo
```

