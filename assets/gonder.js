window.onload = function() {

    // ---  Init ---

    // --- //Init ---



    // --- CKEditor ---

    var editor = CKEDITOR.replace( 'message', {
            filebrowserBrowseUrl: '/assets/filemanager/index.html?config=../../filemanager.config',
            plugins:
            //'dialogui,' +
            'dialog,' +
            'a11yhelp,' +
            'dialogadvtab,' +
            'basicstyles,' +
            'bidi,' +
            'blockquote,' +
            'clipboard,' +
            'button,' +
            'panelbutton,' +
            'panel,' +
            'floatpanel,' +
            'colorbutton,' +
            'colordialog,' +
            'templates,' +
            'menu,' +
            'contextmenu,' +
            'div,' +
            'resize,' +
            'toolbar,' +
            'elementspath,' +
            'enterkey,' +
            'entities,' +
            'popup,' +
            'filebrowser,' +
            'find,' +
            'fakeobjects,' +
            //'flash,' +
            'floatingspace,' +
            'listblock,' +
            'richcombo,' +
            'font,' +
            //'forms,' +
            'format,' +
            'horizontalrule,' +
            'htmlwriter,' +
            //'iframe,' +
            'wysiwygarea,' +
            'image,' +
            'indent,' +
            'indentblock,' +
            'indentlist,' +
            //'smiley,' +
            'justify,' +
            'menubutton,' +
            //'language,' +
            'link,' +
            'list,' +
            'liststyle,' +
            'magicline,' +
            'maximize,' +
            //'newpage,' +
            'pagebreak,' +
            'pastetext,' +
            'pastefromword,' +
            'preview,' +
            'print,' +
            'removeformat,' +
            //'save,' +
            'selectall,' +
            'showblocks,' +
            'showborders,' +
            'sourcearea,' +
            'specialchar,' +
            'scayt,' +
            'stylescombo,' +
            'tab,' +
            'table,' +
            'tabletools,' +
            'undo,' +
            'wsc,' +
            'docprops,',

            allowedContent:  true,
            removeFormatAttributes: '',
        }
    );

    editor.on( 'instanceReady', function() {
        console.log( editor.filter.allowedContent );
    } );

    // --- /CKEditor ---

};

function loadProfiles() {

}

function editProfile(id) {
    $.ajax({
        type: "GET",
        url: "/api/mailer/profile/" + id,
        async: false,
    }).done(function(data) {
        console.log(data)
        $("#profileId").val(data["id"])
        $("#profileName").val(data["name"])
        $("#profileHost").val(data["host"])
        $("#profileIface").val(data["iface"])
        $("#profileServer").val(data["server"])
        $("#profileStreams").val(data["stream"])
        $("#profileDelay").val(data["delay"])
    });
    socksServerInput()
}

function saveProfile() {
    var method
    id = $("#profileId").val()
    if (id == 0) {
        method = "POST"
    } else {
        method = "PUT"
    }
    var data = {
        "id": $("#profileId").val(),
        "name": $("#profileName").val(),
        "iface": $("#profileIface").val(),
        "server": $("#profileServer").val(),
        "host": $("#profileHost").val(),
        "stream": $("#profileStreams").val(),
        "delay": $("#profileDelay").val(),
    };
    $.ajax({
        type: method,
        url: "/api/mailer/profile/",
        async: false,
        data: JSON.stringify(data),
        contentType: "application/json; charset=utf-8",
        dataType: "json",
    }).done(function(data) {
        $('#profile').modal('hide');
        location.reload();
    });
   // $('#profile').modal('hide');
}

function socksServerInput() {
    if ($("#profileIface").val() == 'socks://') {
        $("#profileServer").show()
    } else {
        $("#profileServer").hide()
    }
}
