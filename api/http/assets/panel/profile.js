function refreshProfilesList(selectedProfile){
    var profile;
    $.ajax({
        type: "GET",
        async: false,
        url: '/api/profilelist',
//        contentType: "application/json; charset=utf-8",
        dataType: "json"
    }).done(function(data) {
        profile = data;
    });
    w2ui['parameter'].set('campaignProfileId', { options: { items: profile } });
    w2ui['parameter'].record['campaignProfileId'] = selectedProfile;
    w2ui['parameter'].refresh();
}

// --- Profile table ---
$('#profile').w2grid({
    name: 'profile',
    header: w2utils.lang('Profiles'),
    show: {
        header: true,
        toolbar: true,
        footer: true,
        toolbarSave: true,
        toolbarDelete: true,
        toolbarAdd: true,
        toolbarSearch: false
    },
    columns: [
        { field: 'recid', caption: w2utils.lang('Id'), size: '50px', style: 'background-color: #efefef; border-bottom: 1px solid white; padding-right: 5px;', attr: "align=right" },
        { field: 'name', caption: w2utils.lang('Name'), size: '30%', editable: { type: 'text' }},
        { field: 'host', caption: w2utils.lang('Host'), size: '20%', editable: { type: 'text' }},
        { field: 'iface', caption: w2utils.lang('Interface'), size: '20%',  editable: { type: 'text' }},
        { field: 'stream', caption: w2utils.lang('Stream'), render: 'int', size: '10%', editable: { type: 'int' }},
        { field: 'resend_delay', caption: w2utils.lang('Resend delay'), render: 'int', size: '10%', editable: { type: 'int' }},
        { field: 'resend_count', caption: w2utils.lang('Resend count'), render: 'int', size: '10%', editable: { type: 'int' }}

    ],
    method: 'GET',
    onAdd: function(event) {
        console.log(event);
        $.ajax({
            url: "api/profiles",
            type: "GET",
            dataType: "json",
            data: {"request": JSON.stringify({"cmd": "add", "group": parseInt(w2ui['group'].getSelection()[0])})},
        }).done(function(data) {
            if (data['status'] == 'error') {
                w2alert(w2utils.lang(data["message"]), w2utils.lang('Error'));
            } else {
                console.log(data);
                w2ui.profile.add({ recid: data.recid });
            }
        }).complete(function(data) {

        });
    }
});
// --- /Profile table ---