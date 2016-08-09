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

$(function () {
    initializeFilters();
    attachSpecFilter();
    attachScenarioToggle();
});
