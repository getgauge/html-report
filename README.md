
html-report
==========

This is the [html-report plugin](http://getgauge.io/documentation/plugins/README.html) for [gauge](http://getgauge.io).


Compiling
---------

````
go run build/make.go
````

For cross platform compilation

````
go run build/make.go --all-platforms
````

Installing
----------
After installing gauge

````
go run build/make.go --install
````

Installing to a CUSTOM_LOCATION

````
go run build/make.go --install --plugin-prefix CUSTOM_LOCATION
````

Creating distributable
----------------------

Note: Run after compiling

````
go run build/make.go --distro
````

For distributable across platforms os, windows and linux for bith x86 and x86_64

````
go run build/make.go --distro --all-platforms
````

New distribution details need to be updated in the html-report-install.json file in  [gauge plugin repository](https://github.com/getgauge/gauge-repository) for a new verison update.
