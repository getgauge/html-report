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

var gaugeReport = angular.module('gauge_report', ['yaru22.hovercard']).config([
    '$compileProvider',
    function ($compileProvider) {
        $compileProvider.aHrefSanitizationWhitelist(/^\s*(https?|ftp|mailto|chrome-extension|data):/);
    }
]).directive('collapsable', function () {
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
        if(step)    $scope.currentStep = step
    };
    $scope.setCurrentConceptStep = function (step) {
        $scope.currentConceptStep = null;
        if(step)    $scope.currentConceptStep = step
    };
    $scope.setCurrentExecutionResult = function (result) {
        $scope.currentExecutionResult = null;
        if(result)    $scope.currentExecutionResult = result
    };
    $scope.setConcept = function (concept) {
        $scope.isConcept = false
        if(concept){
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

    $scope.formattedTime = function (timeInMs) {
        var sec = Math.floor(timeInMs / 1000);
        timeInMs %= 1000;

        var min = Math.floor(sec / 60);
        sec %= 60;

        var hour = Math.floor(min / 60);
        min %= 60;

        return convertTo2Digit(hour) + ":" + convertTo2Digit(min) + ":" + convertTo2Digit(sec);
    };

    function convertTo2Digit(value) {
        if (value < 10) {
            return "0" + value;
        }
        return value;
    };

    $scope.getScreenshotSrc = function (screenshot) {
        return "data:image/png;base64," + screenshot
    };

    $scope.summaryItems = [
        {"key": "Executed", "value": $scope.result.specResults.length},
        {"key": "Failure", "value": $scope.result.specsFailedCount, failed: true},
        {"key": "Success Rate", "value": $scope.result.successRate + "%"},
        {"key": "Time", "value": $scope.formattedTime($scope.result.executionTime)},
        {"key": "Environment", "value": $scope.result.environment},
        {"key": "Tags", "value": $scope.result.tags}
    ];
});
