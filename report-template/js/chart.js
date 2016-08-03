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
