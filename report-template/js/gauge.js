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

var gaugeReport = angular.module('gauge_report', ['yaru22.hovercard', 'nvd3']).config([
    '$compileProvider',
    function ($compileProvider) {
        $compileProvider.aHrefSanitizationWhitelist(/^\s*(https?|ftp|mailto|chrome-extension|data):/);
    }]).directive('collapsable', function () {
    return function ($scope, $element) {
        $element.bind('click', function () {
            $element.parent().toggleClass("collapsed");
        });
    };
});

function init() {
    (function addIndexOf() {
        if (!Array.prototype.indexOf) {
            Array.prototype.indexOf = function (obj, start) {
                for (var i = (start || 0), j = this.length; i < j; i++) {
                    if (this[i] === obj) {
                        return i;
                    }
                }
                return -1;
            }
        }
    })();
}

gaugeReport.controller('mainController', function ($scope) {
    init();
    $scope.result = gaugeExecutionResult.suiteResult;
    $scope.itemTypesMap = itemTypesMap;
    $scope.paramTypesMap = parameterTypesMap;
    $scope.fragmentTypesMap = fragmentTypesMap;
    $scope.dataTableIndex = 0;
    $scope.hookFailure = null;
    $scope.isPreHookFailure = false;
    $scope.conceptList = [];
    $scope.count = 0;
    $scope.isConcept = false;

    $scope.allPassed = function () {
        return !$scope.result.failed
    };

    $scope.loadSpecification = function (specification) {
        $scope.currentSpec = specification;
    };

    $scope.initializeLightbox = initLightbox;

    $scope.setDataTableIndex = function (index) {
        $scope.dataTableIndex = index
    };

    $scope.isRowFailure = function (index) {
        var failedRows = $scope.currentSpec.failedDataTableRows;
        if (failedRows === undefined)
            return false;
        else
            return failedRows.indexOf(index) != -1
    };

    $scope.setCurrentStep = function (step) {
        $scope.currentStep = null;
        if (step) $scope.currentStep = step
    };

    $scope.setCurrentConceptStep = function (step) {
        $scope.currentConceptStep = null;
        if (step) $scope.currentConceptStep = step
    };

    $scope.setCurrentExecutionResult = function (result) {
        $scope.currentExecutionResult = null;
        if (result) $scope.currentExecutionResult = result
    };

    $scope.setConcept = function (concept) {
        $scope.isConcept = false
        if (concept) {
            $scope.isConcept = true
            $scope.conceptList.push(concept)
            $scope.currentStep = concept.conceptStep
        }
    };

    $scope.getTopConcept = function () {
        return $scope.conceptList.pop()
    };

    $scope.setCurrentScenario = function (scenario) {
        $scope.currentScenario = scenario
    };

    $scope.getFragmentName = function (name) {
        return name || "table"
    };

    $scope.setHookFailure = function (hookFailure) {
        $scope.hookFailure = hookFailure;
    };

    $scope.hookFailureType = function () {
        return $scope.isPreHookFailure ? "Before-hook failure" : "After-hook failure"
    };

    $scope.isNewLine = function (text) {
        return text === "\n";
    };

    $scope.formattedTime = function (timeInMs, prefix) {
        if (timeInMs == undefined) return "";
        var sec = Math.floor(timeInMs / 1000);

        var min = Math.floor(sec / 60);
        sec %= 60;

        var hour = Math.floor(min / 60);
        min %= 60;

        return (prefix || "") + convertTo2Digit(hour) + ":" + convertTo2Digit(min) + ":" + convertTo2Digit(sec);
    };

    function convertTo2Digit(value) {
        if (value < 10) {
            return "0" + value;
        }
        return value;
    }

    $scope.getScreenshotSrc = function (screenshot) {
        return "data:image/png;base64," + screenshot
    };

    $scope.sort = function (items) {
        if (!items) return;
        var passedScenarios = [];
        var failedScenarios = [];
        return items.filter(function (item) {
            if (itemTypesMap[item.itemType] != "Scenario") return true;
            item.scenario.failed ? failedScenarios.push(item) : passedScenarios.push(item);
        }).concat(failedScenarios).concat(passedScenarios);
    };

    $scope.setCurrentSpec = function (isFirst, specResult) {
        if (isFirst)
            $scope.currentSpec = specResult;
    };

    $scope.getStatus = function (step) {
        if (step.stepExecutionResult.skipped)
            return "skipped";
        else if (step.stepExecutionResult.executionResult) {
            return step.stepExecutionResult.executionResult.failed;
        }
        return undefined
    };

    $scope.summaryItems = [{
        "key": "Executed",
        "value": $scope.result.specResults.length
    }, {
        "key": "Failure",
        "value": $scope.result.specsFailedCount,
        failed: true
    }, {
        "key": "Skipped",
        "value": $scope.result.specsSkippedCount,
        skipped: true
    }, {
        "key": "Success Rate",
        "value": $scope.result.successRate + "%"
    }, {
        "key": "Time",
        "value": $scope.formattedTime($scope.result.executionTime)
    }, {
        "key": "Environment",
        "value": $scope.result.environment
    }, {
        "key": "Tags",
        "value": $scope.result.tags
    }];

    $scope.isEmpty = function (item) {
        if (typeof(item) === "string" && item.length <= 0) {
            return true;
        }
        return false;
    };

    var myColors = ["#A3C273", "#FF6969", "#C6C8C1"];
    d3.scale.myColors = function () {
        return d3.scale.ordinal().range(myColors);
    };

    $scope.options = {
        chart: {
            type: 'pieChart',
            height: 300,
            donut: true,
            donutRatio: 0.3,
            x: function (d) {
                return d.label;
            },
            y: function (d) {
                return d.score;
            },
            showLabels: false,
            showValues: true,
            transitionDuration: 500,
            labelThreshold: 0.01,
            color: d3.scale.myColors().range(),
            legend: {
                margin: {
                    top: 5,
                    right: 35,
                    bottom: 5,
                    left: 0
                }
            }
        }
    };

    $scope.data = [
        {
            label: "Passed",
            score: $scope.result.specResults.length
        },
        {
            label: "Failed",
            score: $scope.result.specsFailedCount
        },
        {
            label: "Skipped",
            score: $scope.result.specsSkippedCount
        }
    ];

});
