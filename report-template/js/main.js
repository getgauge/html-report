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

$(function () {
    initializeFilters();
    attachSpecFilter();
    attachScenarioToggle();
    registerHovercards();
    registerConceptToggle();
});
