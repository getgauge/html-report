html-report
==========

 [ ![Download Nightly](https://api.bintray.com/packages/gauge/html-report/Nightly/images/download.svg) ](https://bintray.com/gauge/html-report/Nightly/_latestVersion) [![Build Status](https://travis-ci.org/getgauge/html-report.svg?branch=master)](https://travis-ci.org/getgauge/html-report)

This is the [html-report plugin](http://getgauge.io/documentation/user/current/plugins/html_report_plugin.html) for [Gauge](http://getgauge.io).


Features
-------

-  A comprehensive test results report template prepared in a html
   format providing the overall summary with drill down of the test
   cases executed and effort spent during the testing for each stage and feature.
-  It provides the details for the defects found during the run.
-  It indicates the tests by color code - failed(red), passed(green) and
   skipped(grey).
-  The failure can be analyzed with the stacktrace and
   screenshot(captures unless overwritten not to).
-  The skipped tests can be analyzed with the given reason.
-  [Custom Messages](https://docs.gauge.org/language.html#custom-messages-in-reports) allows users to add messages at runtime.


**Sample HTML Report documemt**

<img src="https://github.com/getgauge/html-report/raw/master/images/sample.png" alt="Create New Project preview" style="width: 600px;"/>

Installation
------------

```
gauge install html-report
```

* Installing specific version
```
gauge install html-report --version 2.1.0
```

#### Offline installation
* Download the plugin from [Releases](https://github.com/getgauge/html-report/releases)
```
gauge install html-report --file html-report-2.1.0-linux.x86_64.zip
```

#### Build from Source

##### Requirements
* [Golang](http://golang.org/)

##### Compiling
Download dependencies
```
go get -t ./...
```

Compilation
```

go run build/make.go
```

For cross-platform compilation

```
go run build/make.go --all-platforms
```

##### Installing
After compilation

```
go run build/make.go --install
```

Installing to a CUSTOM_LOCATION

```
go run build/make.go --install --plugin-prefix CUSTOM_LOCATION
```

#### Creating distributable

Note: Run after compiling

```
go run build/make.go --distro
```

For distributable across platforms: Windows and Linux for both x86 and x86_64

```
go run build/make.go --distro --all-platforms
```

New distribution details need to be updated in the `html-report-install.json` file in the [gauge plugin repository](https://github.com/getgauge/gauge-repository) for a new version update.

Configuration
-------------

The HTML report plugin can be configured by the properties set in the
`env/default.properties` file in the project.

The configurable properties are:

**gauge_reports_dir**

-  Specifies the path to the directory where the execution reports will
   be generated.

-  Should be either relative to the project directory or an absolute
   path. By default it is set to `reports` directory in the project

**overwrite_reports**

-  Set to ``true`` if the reports **must be overwritten** on each
   execution maintaining only the latest execution report.

-  If set to `false` then a _**new report**_ will be generated on each execution in the reports directory in a nested time-stamped directory. By sdefault it is set to `true`.


**GAUGE_HTML_REPORT_THEME_PATH**

-  Specifies the path to the custom theme directory.

-  Should be either relative to the project directory or an absolute
   path. By default, `default` theme shipped with gauge is used.


Report re-generation
-------------------

If report generation fails due to some reason, we don't have to re-run the tests again.

The html-report plugin now generates a last_run_result.json file in the root of the reports directory.
There is also a symlink to the html-report executable in the same location.

**To regenerate the report**

- Navigate to the reports directory
- run ./html-report --input=last_run_result.json --output="/some/path"

**Note:** The output directory is created. Take care not to overwrite an existing directory

While regenerating a report, the default theme is used. A custom can be used if ``--theme`` flag is specified with the path to the custom theme.


License
-------

![GNU Public License version 3.0](http://www.gnu.org/graphics/gplv3-127x51.png)
`html-report` is released under [GNU Public License version 3.0](http://www.gnu.org/licenses/gpl-3.0.txt)

Copyright
---------

Copyright 2015 ThoughtWorks, Inc.
