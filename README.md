html-report
==========

 [ ![Download Nightly](https://api.bintray.com/packages/gauge/html-report/Nightly/images/download.svg) ](https://bintray.com/gauge/html-report/Nightly/_latestVersion)  [![Build Status](https://app.snap-ci.com/getgauge/html-report/branch/master/build_image)](https://app.snap-ci.com/getgauge/html-report/branch/master) [![Build Status](https://travis-ci.org/getgauge/html-report.svg?branch=master)](https://travis-ci.org/getgauge/html-report)

This is the [html-report plugin](http://getgauge.io/documentation/user/current/plugins/html_report_plugin.html) for [Gauge](http://getgauge.io).

Install through Gauge
---------------------
```
gauge --install html-report
```

* Installing specific version
```
gauge --install html-report --plugin-version 2.1.0
```

### Offline installation
* Download the plugin from [Releases](https://github.com/getgauge/html-report/releases)
```
gauge --install html-report --file html-report-2.1.0-linux.x86_64.zip
```

Build from Source
-----------------

### Requirements
* [Golang](http://golang.org/)

### Compiling

```
go run build/make.go
```

For cross-platform compilation

```
go run build/make.go --all-platforms
```

### Installing
After compilation

```
go run build/make.go --install
```

Installing to a CUSTOM_LOCATION

```
go run build/make.go --install --plugin-prefix CUSTOM_LOCATION
```

### Creating distributable

Note: Run after compiling

```
go run build/make.go --distro
```

For distributable across platforms: Windows and Linux for both x86 and x86_64

```
go run build/make.go --distro --all-platforms
```

New distribution details need to be updated in the `html-report-install.json` file in the [gauge plugin repository](https://github.com/getgauge/gauge-repository) for a new version update.

License
-------

![GNU Public License version 3.0](http://www.gnu.org/graphics/gplv3-127x51.png)
`html-report` is released under [GNU Public License version 3.0](http://www.gnu.org/licenses/gpl-3.0.txt)

Copyright
---------

Copyright 2015 ThoughtWorks, Inc.
