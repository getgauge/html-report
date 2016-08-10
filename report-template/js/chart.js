// Copyright 2015 ThoughtWorks, Inc.

// This file is part of getgauge/html-report.

// getgauge/html-report is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// getgauge/html-report is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with getgauge/html-report.  If not, see <http://www.gnu.org/licenses/>.

var createChart = function(passed, failed, skipped) {
    var data = [{ "label": "Passed", "value": passed },
        { "label": "Failed", "value": failed },
        { "label": "Skipped", "value": skipped }
    ];

    var myColors = ["#27caa9", "#e73e48", "#999999"];
    d3.scale.myColors = function() {
        return d3.scale.ordinal().range(myColors);
    };

    nv.addGraph(function() {
        var chart = nv.models.pieChart()
            .height(180)
            .x(function(d) {
                return d.label
            })
            .y(function(d) {
                return d.value
            })
            .showLabels(false)
            .labelThreshold(.01)
            .donut(true)
            .donutRatio(0.2)
            .showLegend(false)
            .color(myColors);

        d3.select(".chart svg")
            .datum(data)
            .transition().duration(350)
            .call(chart);

        return chart;
    });
}
