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

function initializeFilters() {
    if (sessionStorage.FilterStatus) {
        filterSpecList(sessionStorage.FilterStatus);
        $('.spec-filter').each(function() {
            if ($(this).data('status')===sessionStorage.FilterStatus) {
                $(this).addClass('active');
            }
        });
    }
    else {
        $('.total-specs').addClass('active');
    }
}

function showFirstSpecContent() {
    $('li.spec-name:visible:first').click();
    if($('li.spec-name:visible:first').length===0){
        $('#specificationContainer').hide();
    }
}

function attachScenarioToggle() {
    $('.row-selector').click(function() {
        $('.row-selector').each(function() { $(this).removeClass('selected');});
        $(this).addClass('selected');
        var tr=$(this).data('rowindex');
        $(".scenario-container").each(function(){
            if($(this).data('tablerow')===tr) { $(this).show();} else {$(this).hide();}
        });
    });
}

function filterSpecList(status) {
    $('#listOfSpecifications li.spec-name').each(function() {
        if($(this).hasClass(status)) {
            $(this).show();
        }
        else {
            $(this).hide();
        }
    });
}

function attachSpecFilter() {
    var resetState = function() {
        $('.spec-filter, .total-specs').each(function(){
            $(this).removeClass('active');}
        );
    };
    $('.spec-filter').click(function(){
        resetState();
        var status = $(this).data('status');
        sessionStorage.FilterStatus = status;
        filterSpecList(status);
        showFirstSpecContent();
        $(this).addClass('active');
    });
    $('.total-specs').click(function () {
        resetState();
        $('#listOfSpecifications li.spec-name').each(function() {
            $(this).show();
        });
        sessionStorage.removeItem('FilterStatus');
        showFirstSpecContent();
        $(this).addClass('active');
    });
}

function registerHovercards() {
    $('span.hoverable').mouseenter(function(e) {
        $(this).next('.hovercard').css({top: e.clientY + 10, left: e.clientX +10}).delay(100).fadeIn();
    }).mouseleave(function() {
        $(this).next('.hovercard').delay(100).fadeOut('fast');
    });
}

function registerConceptToggle() {
    $('.concept').click(function() {
        var conceptSteps = $(this).next('.concept-steps');
        var iconClass = $(conceptSteps).is(':visible') ? "plus" : "minus";
        $(conceptSteps).fadeToggle('fast', 'linear');
        $(this).find("i.fa").removeClass("fa-minus-square").removeClass("fa-plus-square").addClass("fa-"+iconClass+"-square");
    });
}

function registerSearch() {
    $('#searchSpecifications').change(function() {
        if(!index) return;
        searchText = $(this).val();
        tagMatches = index.tags[searchText];
        specMatches=[];
        for(var spec in index.specs) {
            if(index.specs.hasOwnProperty(spec) && spec.startsWith(searchText)) {
                index.specs[spec].forEach(function(x) {specMatches.push(x)});
            }
        }
        $(".spec-list a").each(function() {
            var href=$(this).attr('href');
            var existsIn = function(arr) {
                return typeof arr !== 'undefined' && $.inArray(href, arr)>=0;
            }
            if(existsIn(tagMatches) || existsIn(specMatches))
                $(this).show();
            else
                $(this).hide();
        })
    });
}

function registerSearchAutocomplete() {
    new autoComplete({
        selector: 'input[id="searchSpecifications"]',
        minChars: 1,
        source: function(term, suggest){
            term = term.toLowerCase();
            var tagChoices = Object.keys(index.tags);
            var specChoices = Object.keys(index.specs);
            var suggestions = [];
            var suggestionPredicate = function(type) { 
                return function(x) {
                    if(x.toLowerCase().startsWith(term))
                        suggestions.push([x, type]);
                }
            };
            tagChoices.forEach(suggestionPredicate("tag"));
            specChoices.forEach(suggestionPredicate("spec"));
            suggest(suggestions);
        },
        renderItem: function (item, search){
            iconClass = item[1] == "tag" ? "tags" : "bars"
            return '<div class="autocomplete-suggestion" data-value="'+ item[0] +'"><i class="fa fa-' + iconClass + '" aria-hidden="true"></i>&nbsp;' + item[0] + '</div>';
        },
        onSelect: function(e, term, item){
            $('#searchSpecifications').val($(item).data('value'));
        }
    });
}

$(function () {
    initializeFilters();
    attachSpecFilter();
    attachScenarioToggle();
    registerHovercards();
    registerConceptToggle();
    registerSearch();
    registerSearchAutocomplete();
});