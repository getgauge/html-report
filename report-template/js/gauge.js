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

marked.setOptions({ gfm: true, sanitize: true, tables: true, breaks: true, smartLists: true });

// Override lightbox listener if it exists
if (typeof listenKey === "function") {
    //
    // listenKey()
    //
    listenKey = function() { document.onkeydown = getKey; };
}

var gaugeReport = angular.module('gauge_report', ['yaru22.hovercard', 'nvd3', 'ngSanitize']).config([
    '$compileProvider',
    function($compileProvider) {
        $compileProvider.aHrefSanitizationWhitelist(/^\s*(https?|ftp|mailto|chrome-extension|data|file):/);
    }
]).config(['$provide', function($provide) {
    $provide.decorator('$sniffer', ['$delegate', function($delegate) {
        $delegate.history = false;
        return $delegate;
    }]);

}]).directive('collapsable', function() {
    return function($scope, $element) {
        $element.bind('click', function() {
            $element.parent().toggleClass("collapsed");
        });
    };
});

function initialize() {
    (function addIndexOf() {
        if (!Array.prototype.indexOf) {
            Array.prototype.indexOf = function(obj, start) {
                for (var i = (start || 0), j = this.length; i < j; i++) {
                    if (this[i] === obj) {
                        return i;
                    }
                }
                return -1;
            };
        }
    })();
}

gaugeReport.controller('mainController', function($scope, $timeout) {
    initialize();
    $scope.result = gaugeExecutionResult.suiteResult;
    $scope.itemTypesMap = itemTypesMap;
    $scope.paramTypesMap = parameterTypesMap;
    $scope.fragmentTypesMap = fragmentTypesMap;
    $scope.dataTableIndex = 0;
    $scope.hookFailure = null;
    $scope.isPreHookFailure = false;
    $scope.isPostHookFailure = false;
    $scope.conceptList = [];
    $scope.count = 0;
    $scope.isConcept = false;
    $scope.tearDownSteps = [];
    $scope.currentMessage = "";
    $scope.search = { query: "", specList: [], timer: undefined, disabled: false, lastSpec: undefined };

    $scope.allPassed = function() {
        return !$scope.result.failed;
    };

    $scope.loadSpecification = function(specification) {
        $scope.currentSpec = $scope.search.lastSpec = specification;
        $scope.currentMessage = "";
    };

    $scope.initializeLightbox = initLightbox;

    $scope.setDataTableIndex = function(index) {
        $scope.dataTableIndex = index;
    };

    $scope.isRowFailure = function(index) {
        var failedRows = $scope.currentSpec.failedDataTableRows;
        if (failedRows === undefined)
            return false;
        else
            return failedRows.indexOf(index) != -1;
    };

    $scope.setCurrentStep = function(step) {
        $scope.currentStep = null;
        if (step) $scope.currentStep = step;
    };

    $scope.setCurrentConceptStep = function(step) {
        $scope.currentConceptStep = null;
        if (step) $scope.currentConceptStep = step;
    };

    $scope.setCurrentExecutionResult = function(result) {
        $scope.currentExecutionResult = null;
        if (result) $scope.currentExecutionResult = result;
    };

    $scope.setConcept = function(concept) {
        $scope.isConcept = false;
        if (!!concept) {
            $scope.isConcept = true;
            $scope.conceptList.push(concept);
            $scope.setCurrentConceptStep(concept.conceptStep);
        }
    };

    $scope.getTopConcept = function() {
        return $scope.conceptList.pop();
    };

    $scope.setCurrentScenario = function(scenario) {
        $scope.currentScenario = scenario;
    };

    $scope.getFragmentName = function(name) {
        return name || "table";
    };

    $scope.setHookFailure = function(hookName, hookFailure) {
        $scope.currentHookName = hookName;
        $scope.hookFailure = hookFailure;
    };

    $scope.hookFailureType = function() {
        return $scope.isPreHookFailure ? "Before-hook failure" : "After-hook failure";
    };

    $scope.isNewLine = function(text) {
        return text === "\n";
    };

    $scope.formattedTime = function(timeInMs, prefix) {
        if (timeInMs === undefined) return "";
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

    $scope.getScreenshotSrc = function(screenshot) {
        return "data:image/png;base64," + screenshot;
    };

    $scope.sort = function(items) {
        if (!items) return true;
        var passedScenarios = [];
        var failedScenarios = [];
        return items.filter(function(item) {
            if (itemTypesMap[item.itemType] !== "Scenario") return true;
            if (item.scenario.failed) {
                failedScenarios.push(item);
            } else {
                passedScenarios.push(item);
            }
            return false;
        }).concat(failedScenarios).concat(passedScenarios);
    };

    $scope.filteredListOfSpecs = $scope.result.specResults;

    $scope.showFailedSpecs = function() {
        if ($scope.isPreHookFailure) return;
        var specs = [];
        angular.forEach($scope.result.specResults, function(specRes) {
            if (specRes.failed) {
                specs.push(specRes);
            }
        });
        if (specs.length > 0) {
            $scope.setCurrentSpec(true, specs[0]);
            $scope.currentMessage = "";
        } else {
            $scope.currentSpec = undefined;
            $scope.currentMessage = "No failed specifications.";
        }
        $scope.filteredListOfSpecs = specs;
    };

    $scope.showPassedSpecs = function() {
        if ($scope.isPreHookFailure) return;
        var specs = [];
        angular.forEach($scope.result.specResults, function(specRes) {
            if (!specRes.failed && specRes.scenarioSkippedCount < specRes.scenarioCount) {
                specs.push(specRes);
            }
        });
        if (specs.length > 0) {
            $scope.loadSpecification(specs[0]);
            $scope.currentMessage = "";
        } else {
            $scope.currentSpec = undefined;
            $scope.currentMessage = "No passed specifications.";
        }
        $scope.filteredListOfSpecs = specs;
    };

    $scope.showSkippedSpecs = function() {
        if ($scope.isPreHookFailure) return;
        var specs = [];
        angular.forEach($scope.result.specResults, function(specRes) {
            if (specRes.skipped && specRes.scenarioCount === specRes.scenarioSkippedCount) {
                specs.push(specRes);
            }
        });
        if (specs.length > 0) {
            $scope.loadSpecification(specs[0]);
            $scope.currentMessage = "";
        } else {
            $scope.currentSpec = undefined;
            $scope.currentMessage = "No skipped specifications.";
        }
        $scope.filteredListOfSpecs = specs;
    };

    $scope.showAllSpecs = function() {
        if ($scope.isPreHookFailure) return;
        $scope.loadSpecification($scope.result.specResults[0]);
        $scope.currentMessage = "";
        $scope.filteredListOfSpecs = $scope.result.specResults;
    };

    $scope.setCurrentSpec = function(isFirst, specResult) {
        if (isFirst && specResult.failed)
            $scope.currentSpec = specResult;
    };

    $scope.addTearDownSteps = function(steps) {
        $scope.tearDownSteps = steps;
    };

    $scope.getStatus = function(step) {
        if (step.stepExecutionResult.skipped)
            return "skipped";
        if (step.stepExecutionResult.executionResult)
            return step.stepExecutionResult.executionResult.failed;
        return undefined;
    };

    $scope.summaryItems = [{
        "key": "Environment",
        "value": $scope.result.environment
    }, {
        "key": "Tags",
        "value": $scope.result.tags
    }, {
        "key": "Success Rate",
        "value": $scope.result.successRate + "%"
    }, {
        "key": "Total Time",
        "value": $scope.formattedTime($scope.result.executionTime)
    }, {
        "key": "Generated On",
        "value": $scope.result.timestamp
    }];

    $scope.isEmpty = function(item) {
        return !!(typeof(item) === "string" && item.length <= 0);
    };

    var searchEnd = function () {
        $scope.search.disabled = true;
         if ($scope.search.query.length === 0) {
             $scope.currentMessage = "";
             if ($scope.search.lastSpec !== undefined) {
                 $scope.loadSpecification($scope.search.lastSpec);
             } else {
                 $scope.loadSpecification($scope.search.specList[0]);
             }
        } else if ($scope.search.specList.length >= 1) {
            if ($scope.currentSpec !== undefined) {
                var matched = $scope.search.specList.filter(function (s) {
                    return s.protoSpec.specHeading === $scope.currentSpec.protoSpec.specHeading;
                });
                if (matched.length < 1) {
                    $scope.loadSpecification($scope.search.specList[0]);
                }
            } else {
                $scope.loadSpecification($scope.search.specList[0]);
            }
        } else if ($scope.search.query.length >= 1 && $scope.search.specList.length === 0) {
            $scope.currentSpec = undefined;
            $scope.currentMessage = "No results found.";
        }
        $scope.search.disabled = false;
    };

    $scope.$watch("search.query", function () {
        if ($scope.search.query.length === 0) {
            searchEnd();
            return;
        }
        if ($scope.search.timer) $timeout.cancel($scope.search.timer);
        $scope.search.timer = $timeout(searchEnd, 200);
    }, true);

    $scope.searchItems = function(searchQuery) {
        return function(spec) {
            if (!searchQuery) return true;
            if (spec.protoSpec.specHeading.toLowerCase().indexOf(searchQuery.toLowerCase()) > -1) return true;
            var tagMatches = spec.protoSpec.items.filter(function(item) {
                var searchList = [];
                if (item.scenario) searchList.push(item.scenario.scenarioHeading);
                if (item.scenario && item.scenario.tags) searchList = searchList.concat(item.scenario.tags);
                if (item.tags && item.tags.tags) searchList = searchList.concat(item.tags.tags);
                return searchList.join(" ").toLowerCase().indexOf(searchQuery.toLowerCase()) > -1;
            });
            if (tagMatches.length) return true;
            return false;
        };
    };

    $scope.showScenario = function(item) {
        if (!$scope.search.query) return true;
        if ($scope.currentSpec.protoSpec.specHeading.toLowerCase().indexOf($scope.search.query.toLowerCase()) > -1) return true;
        if ($scope.currentSpec.protoSpec.tags && $scope.currentSpec.protoSpec.tags.join(" ").toLowerCase().indexOf($scope.search.query.toLowerCase()) > -1) return true;
        if (item.contexts && item.contexts.length) {
            var matchedContexts = item.contexts.filter(function(context) {
                if (context.step && context.step.parsedText) {
                    return context.step.parsedText.indexOf($scope.search.query.toLowerCase()) > -1;
                }
                return false;
            });
            if (matchedContexts.length > 0) return true;
        }
        if (item.scenarioHeading.toLowerCase().indexOf($scope.search.query.toLowerCase()) < 0) {
            if (item.tags) return item.tags.join(" ").toLowerCase().indexOf($scope.search.query.toLowerCase()) > -1;
        } else {
            return true;
        }
        return false;
    };

    var myColors = ["#27caa9", "#e73e48", "#999999"];
    d3.scale.myColors = function() {
        return d3.scale.ordinal().range(myColors);
    };

    $scope.options = {
        chart: {
            type: 'pieChart',
            height: 200,
            margin: {
                top: 0,
                right: 0,
                bottom: 0,
                left: 0
            },
            donut: true,
            donutRatio: 0.2,
            x: function(d) {
                return d.label;
            },
            y: function(d) {
                return d.score;
            },
            showLabels: false,
            showValues: true,
            transitionDuration: 500,
            labelThreshold: 0.01,
            color: d3.scale.myColors().range(),
            showLegend: false,
            valueFormat: d3.format("d")
        }
    };

    if ($scope.result && $scope.result.specResults) {
        $scope.totalSpecs = $scope.result.specResults.length || 0;
        $scope.failed = $scope.result.specsFailedCount || 0;
        $scope.skipped = $scope.result.specsSkippedCount || 0;
        $scope.passed = $scope.totalSpecs - ($scope.failed + $scope.skipped);
        $scope.data = [{
            label: "Passed",
            score: $scope.passed
        }, {
            label: "Failed",
            score: $scope.failed
        }, {
            label: "Skipped",
            score: $scope.skipped
        }];
    } else if ($scope.result && $scope.result.preHookFailure) {
        $scope.totalSpecs = $scope.failed = $scope.skipped = $scope.passed = 0;
        $scope.hookFailure = $scope.result.preHookFailure;
        $scope.isPreHookFailure = true;
        $scope.data = [{
            label: "Passed",
            score: 0
        }, {
            label: "Suite Failed",
            score: 1
        }, {
            label: "Skipped",
            score: 0
        }];
    }
    if ($scope.result && $scope.result.postHookFailure) {
        $scope.isPostHookFailure = true;
    }

    $scope.projectName = $scope.result.projectName;

    $scope.parseComment = function(item) {
        return marked(item.comment.text);
    };


    $scope.isStepFailure = function(result) {
        return result.executionResult && result.executionResult.failed && result.executionResult.errorMessage && result.executionResult.stackTrace;
    };

    $scope.formatMessage = function (msg) {
        return msg.replace(/\n/g, "<br/>").replace(/\s/g, "&nbsp;");
    };
});
