var intervalRequest=5000;
var timeResendRequest=90;
var sendRequest=false;
var fnsEmptyResponseCount=1;
function actionResponse(a)
{
    if(a.status=="success"){
        if(a.url){
            stopIntervalSend();
            $(location).attr("href",a.url)
        }
    }
    return a.status
}

function completeTime(a)
{
    return Math.round(+new Date()/1000)>=a
}

function stopIntervalSend()
{
    if(sendRequest){
        window.clearInterval(sendRequest)
    }
}

function splitData(e)
{
    var c=decodeURI(e).split("&");
    var f=[];
    for(var b=0;b<c.length;b++){
        var d=c[b].split("=");var a=d[0].replace("[]","");
        if(f[a]==undefined){
            f[a]=""
        }
        f[a]+=d[1]
    }
    return f
}

function createSearchData(b,c)
{
    var a=$("#poisk-inprocess");
    a.find("div").addClass("hidden");
    a.find(".proccessFind").removeClass("hidden");
    stopIntervalSend();
    $.ajax(
        {
            url:createRequestSearchUrl,
            type:"post",
            data:{data:b,article:c},
            success:function(d){
                fnsEmptyResponseCount=1;
                var f=$("#poisk-inprocess");
                if(d.status=="success"){
                    var e=Math.round(+new Date()/1000)+timeResendRequest;
                    sendRequest=setInterval(
                        function(){
                            var i=getCharges(d.uin);
                            var g=actionResponse(i);
                            if(g=="success"&&i.url){
                                return
                            }
                            if(g=="fail"&&!completeTime(e)){

                            }else{
                                if(g=="fail"&&completeTime(e)){
                                    f.find("div").addClass("hidden");
                                    f.find(".errorFkResponse").removeClass("hidden");
                                    yaMetricTarget("fk_search_error_response");
                                    stopIntervalSend()
                                }else{
                                    if(g==false){
                                        showErrorMessage();
                                        yaMetricTarget("fk_search_error");
                                        stopIntervalSend()
                                    }else{
                                        if(g=="success"&&c=="gibdd"){
                                            var h=splitData(b);
                                            if(h["gibdd-search"]=="licreg"){
                                                if(h.license.length&&h.regnum.length){
                                                    $(".emptyResponseText").text("Информация о задолженностях может отсутствовать в связи с задержкой поступления информации о начислениях в государственную систему ГИС ГМП.");
                                                    showEmptyMessage()
                                                }else{
                                                    if(h.license.length||h.regnum.length){
                                                        $(".emptyResponseText").text("Попробуйте повторить поиск, указав данные водительского удостоверения и свидетельства о регистрации ТС.");
                                                        showEmptyMessage()
                                                    }
                                                }
                                            }
                                            yaMetricTarget("search_empty");
                                            stopIntervalSend()
                                        }else{
                                            if(g=="success"&&c=="ufns"){
                                                if(fnsEmptyResponseCount==3){
                                                    h=splitData(b);
                                                    if(h["fns-search"]=="inn"&&!h.docno){
                                                        $(".emptyResponseText").html("Информация об имеющихся задолженностях может отсутствовать в связи с задержкой поступления информации о начислениях в государственную систему ГИС ГМП.<br/>Для оплаты налогов по квитанции воспользуйтесь поиском по индексу налогового документа.")
                                                    }
                                                    yaMetricTarget("search_empty");
                                                    showEmptyMessage();
                                                    stopIntervalSend()
                                                }
                                                fnsEmptyResponseCount++
                                            }else{
                                                if(g=="success"){yaMetricTarget("search_empty");
                                                    showEmptyMessage();
                                                    stopIntervalSend()
                                                }else{
                                                    stopIntervalSend()
                                                }
                                            }
                                        }
                                    }
                                }
                            }
                        },intervalRequest)
                }else{
                    if(d.status==false){
                        showErrorMessage();
                        yaMetricTarget("fk_search_error")
                    }
                }
            }

            ,error:function(){
                showErrorMessage();
            yaMetricTarget("fk_search_error")
            }
        })
}

function showEmptyMessage(){
    var a=$("#poisk-inprocess");
    a.find("div").addClass("hidden");
    a.find(".emptyResponse").removeClass("hidden")
}

function showErrorMessage(){
    var a=$("#poisk-inprocess");
    a.find("div").addClass("hidden");
    a.find(".errorResponse").removeClass("hidden")
}

function getCharges(b){
    var a=false;
    $.ajax(
        {
            url:getChargesUrl,
            async:false,
            type:"post",
            data:{uin:b},
            success:function(c){a=c}
        });
    return a
}

$(function(){
    var a=$("input[name=ufnsuin]").val();
    if(a){var b=$("#poisk-inprocess");
        b.find("div").addClass("hidden");
        b.find(".proccessFind").removeClass("hidden");
        b.modal();
        $.ajax({
            url:urlSearchData,type:"post",data:{
                uin:a
            },
            success:function(c,e,d){
                if(c.status=="success"){
                    window.location.replace(c.data);
                    window.location.href(c.data)
                }else{
                    showEmptyMessage();
                    $("#poisk-inprocess").find("div a.btn-link").hide()
                }
            },
            error:function(c,e,d){
                showErrorMessage();
                $("#poisk-inprocess").find("div a.btn-link").hide()
            }
        })
    }
});

$(function(){$("#poisk-inprocess").on("hidden",function(){stopIntervalSend()})});


var docno_attempts=[1,2,3,4,5,6,7,8,9,10,1,2,3,4,5,6,7,8,9,10,1],
    docno_attempts25=[1,2,3,4,5,6,7,8,9,10,1,2,3,4,5,6,7,8,9,10,1,2,3,4,5,6],
    inn_attempts_penult=[7,2,4,10,3,5,9,4,6,8],
    inn_attempts_last=[3,7,2,4,10,3,5,9,4,6,8],
    snils_attempts=[9,8,7,6,5,4,3,2,1],
    inn_check_number=11,
    docno_check_number=11,
    docno_last_rank=19,
    docno_last_rank25=24,
    snils_last_rank=9,
    inn_n2_rank=10,
    inn_n1_rank=11,
    inn_length=12,
    docno_length=20,
    docno_length25=25,
    snils_length=11;

function calcCheckSum(d,a,f,b){
    var g=0;
    var c=a;
    for(var e=0;e<b;e++){
        g+=f[c]*d[e];c++
    }
    return g
}

function splitDocNo(c)
{
    var e,d;
    if(c.length==docno_length){
        e=docno_last_rank;
        d=docno_attempts
    }else{
        if(c.length==docno_length25){
            e=docno_last_rank25;
            d=docno_attempts25
        }
    }
    var b=0,
        a=calcCheckSum(c,b,d,e)%docno_check_number;
    if(a==10&&b==0){
        b=2;
        a=calcCheckSum(c,b,d,e)%docno_check_number
    }
    if(a==10&&b==2){
        a=0
    }
    return a==c[e]&&parseInt(c)!=0
}

function splitInn(d){
    var a=0,
        b=calcCheckSum(d,a,inn_attempts_penult,inn_n2_rank)%inn_check_number;
    b=b!="10"?b:"0";
    if(b!=d[inn_n2_rank]){
        return false
    }
    var c=calcCheckSum(d,a,inn_attempts_last,inn_n1_rank)%inn_check_number;
    c=c!="10"?c:"0";
    return c==d[inn_n1_rank]
}
function splitSnils(d){
    var b=0,a,c=calcCheckSum(d,b,snils_attempts,snils_last_rank);
    if(c<100){
        a=c
    }else{
        if(c==100||c==101){
            a="00"
        }else{
            if(c>101){
                c=c%101;
                if(c<100){a=c}else{if(c==100||c==101){a="00"}}
            }
        }
    }
    return a==d.substr(-2,2)
}

function validateDocNo(){
    var a=$("#newDocNo1");
    return validateDocNoByObject(a)
}

function validateDocNoByObject(d){
    var a=false;
    var b=d.val();
    if(b.length==docno_length||b.length==docno_length25){
        a=splitDocNo(b)
    }
    var e=d.parents(".control-group");
    var c=e.find(".errormark");
    c.html(d.attr("data-format-message"));
    if(!a){
        if(!e.hasClass("error")){e.addClass("error")}}else{e.removeClass("error")}
        return a
}

function validateInn(c){
    var a=false;
    var e=c.val();
    if(e.length==0){return true}
    if(e.length==inn_length){a=splitInn(e)}
    var d=c.parents(".control-group");
    var b=d.find(".errormark");
    b.html("Неправильно введён ИНН");
    if(!a){
        if(!d.hasClass("error")){d.addClass("error")}
    }else{
        d.removeClass("error")
    }
    return a
}

function validateSnils(c){var a=false;var d=c.val();if(d.length==0){return true}if(parseInt(d)==0){return false}if(d.substring(0,9)<=minSnilsNum){return true}if(d.length==snils_length){a=splitSnils(d)}var e=c.parents(".control-group");var b=e.find(".errormark");b.html("Неправильно введён СНИЛС");if(!a){if(!e.hasClass("error")){e.addClass("error")}}else{e.removeClass("error")}return a}

function ajaxValidateInn(a){var b=true;$.ajax({url:validateInnUrl,type:"POST",async:false,data:{inn:a.val()},success:function(c){var e=c.result;if(e==false){var f=a.parents(".control-group");f.addClass("error");var d=f.find(".errormark");d.html("Неправильно введён ИНН");if(!f.hasClass("error")){f.addClass("error")}b=false}},error:function(c){}});return b}
function ajaxValidateSnils(c,a){var b=true;$.ajax({url:validateSnilsUrl,type:"POST",async:false,data:{snils:a},success:function(d){var f=d.result;if(f==false){var g=c.parents(".control-group");g.addClass("error");var e=g.find(".errormark");e.html("Неправильно введён СНИЛС");if(!g.hasClass("error")){g.addClass("error")}b=false}},error:function(d){}});return b}
function fnsValidateInn(){return validateInn($("#fnsInn"))}
function fsspValidateInn(){return validateInn($("#fsspInn"))}
function fsspOldDocNumValidate(){var c="^\\d{1,7}/\\d{2}/\\d{2}/\\d{2}$";var a=$(".masked_ip1").val().match(new RegExp(c,"m"));if(!a){var d=$(".masked_ip1").parents(".control-group");var b=d.find(".errormark");b.html("Неверный формат");d.addClass("error")}return a!=null}
function fsspNewDocNumValidate(){var c="^\\d{1,7}/\\d{2}/\\d{5}$";var a=$(".masked_ip2").val().match(new RegExp(c,"m"));if(!a){var d=$(".masked_ip2").parents(".control-group");var b=d.find(".errormark");b.html("Неверный формат");d.addClass("error")}return a}
function validatePayerName(){var c=["fio1"];var e=declinePayerName.replace(/,/g,"|");var b=new RegExp("(^|-| )("+e+")($|-| )","i");for(var a=0;a<c.length;a++){var f=$("input[name="+c[a]+"]");var d=f.val();if(b.exec(d)!=null){$("div.userfio-block").addClass("error").find("div.errormark").text("Необходимо указать ФИО плательщика");return false}}return true}
function validateRRCode(){var b=$("#rrCode");var a=false;$.ajax({type:"POST",url:validateRRCodeUrl,data:{code:b.val()},async:false,success:function(d){a=d.result;var e=b.parents(".control-group");var c=e.find(".errormark");if(!a&&!e.hasClass("error")){c.html(errorMessageFromRRCode(d.error_code));e.addClass("error")}else{e.removeClass("error")}},error:function(c){}});return a}
function validateAmountUfms(){var a=$("#ufms-summa1");return checkMaxAmount(a)}
function validateAmount(){var a=$("#amountField");return checkMaxAmount(a)}
function checkMaxAmount(c,a){a=typeof a!=="undefined"?a:max_pay_amount;a=parseInt(a);var b=aggregateValFromInputBlocks(c.parent(),"%1.%2");b=parseFloat(b);if(b>a){c.closest("div.summ-block").addClass("error").find("div.errormark").text("Сумма одного платежа не должна превышать "+a+" рублей");return false}return true}
function errorMessageFromRRCode(a){switch(a.toString()){case"2":return"Истёк допустимый для оплаты срок";case"3":case"4":return"Ошибка при обработке введенного кода платежа.";case"5":return"Оплата по указанному коду была произведена ранее.";case"6":return"Оплата услуг Росреестра доступна только для физических лиц.";default:return"Некорректно введены данные"}}function checkPayerDocTypes(){var b=$("#passport");var a=$("#snils");if(b.prop("checked")){b.closest("li").find(".controls input[type=text]").prop("disabled",false);a.closest("li").find(".controls input[type=text]").prop("disabled",true);a.closest("li").find(".controls input[type=text]").prop("required",false);b.closest("li").find(".controls input[type=text]").prop("required",true)}else{if(a.prop("checked")){b.closest("li").find(".controls input[type=text]").prop("disabled",true);a.closest("li").find(".controls input[type=text]").prop("disabled",false);a.closest("li").find(".controls input[type=text]").prop("required",true);b.closest("li").find(".controls input[type=text]").prop("required",false)}}}function isPassportInvalid(a){return !a.match(/^((\d{4})(\d{6}))$/)||parseInt(a)==0}
function isRegNumInvalid(a){var b=/^(\d{2}(([а-яА-Я]{2})|(\d{2}))\d{6})?$/;return !a.match(b)||parseInt(a)==0};