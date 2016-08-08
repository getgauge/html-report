$(function() {
$('.row-selector').click(function() {
    $('.row-selector').each(function() { $(this).removeClass('selected');});
    $(this).addClass('selected');
    var tr=$(this).data('rowindex');
    console.log("1 : "+ tr)
    $(".scenario-container").each(function(){
        console.log("2 : "+ $(this).data('tablerow'))
        if($(this).data('tablerow')===tr) { $(this).show();} else {$(this).hide();}
    });
});});
