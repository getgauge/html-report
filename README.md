html-report
==========

 [ ![Download Nightly](https://api.bintray.com/packages/gauge/html-report/Nightly/images/download.svg) ](https://bintray.com/gauge/html-report/Nightly/_latestVersion)


This is the [html-report plugin](http://getgauge.io/documentation/user/current/plugins/README.html) for [gauge](http://getgauge.io).

Install through Gauge
---------------------
````
gauge --install html-report
````

* Installing specific version
```
gauge --install html-report --plugin-version 1.0.1
```

### Offline installation
* Download the plugin from [Releases](https://github.com/getgauge/html-report/releases)
```
gauge --install html-report --file html-report-1.0.1-linux.x86_64.zip
```

Build from Source
-----------------

### Requirements
* [Golang](http://golang.org/)

### Compiling

````
go run build/make.go
````

For cross platform compilation

````
go run build/make.go --all-platforms
````

### Installing
After compilation

````
go run build/make.go --install
````

Installing to a CUSTOM_LOCATION

````
go run build/make.go --install --plugin-prefix CUSTOM_LOCATION
````

### Creating distributable

Note: Run after compiling

````
go run build/make.go --distro
````

For distributable across platforms os, windows and linux for bith x86 and x86_64

````
go run build/make.go --distro --all-platforms
````

New distribution details need to be updated in the html-report-install.json file in  [gauge plugin repository](https://github.com/getgauge/gauge-repository) for a new verison update.

License
-------

![GNU Public License version 3.0](http://www.gnu.org/graphics/gplv3-127x51.png)
Html-Report is released under [GNU Public License version 3.0](http://www.gnu.org/licenses/gpl-3.0.txt)

Copyright
---------

Copyright 2015 ThoughtWorks, Inc.
