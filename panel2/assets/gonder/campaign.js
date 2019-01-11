// --- Campaign table ---
$('#campaign').w2grid({
    name: 'campaign',
    header: w2utils.lang('Campaign'),
    keyboard : false,
    show: {
        header: true,
        toolbar: true,
        footer: true,
        toolbarDelete: false,
        toolbarAdd: true,
        toolbarSave: true,
        toolbarSearch: true
    },
    columns: [
        { field: 'recid', caption: w2utils.lang('Id'), size: '50px', sortable: true, attr: "align=right" },
        { field: 'name', caption: w2utils.lang('Name'), size: '100%', sortable: true, editable: { type: 'text' } }
    ],
    multiSelect: false,
    sortData: [{ field: 'recid', direction: 'DESC' }],
    url: '/api/campaigns',
    method: 'GET',
    toolbar: {
        items: [
            {id: 'clone', type: 'button', caption: w2utils.lang('Clone'), icon: 'w2ui-icon-columns'}
        ],

        onClick: function (event) {
            if (event.target == 'clone')
            {
                cloneCampaign(parseInt(w2ui['campaign'].getSelection()[0]));
            }
        }
    },

    onAdd: function (event) {
        addCampaign(parseInt(w2ui['group'].getSelection()[0]));
    },

    onSelect: function (event) {
        var record = this.get(event.recid);
        getCampaign(record.recid, record.name)
    },

    onSave: function(event) {
        //console.log(event);
    }
});
// --- /Campaign table ---

function cloneCampaign(campaignId) {
    if (isNaN(campaignId)) {
        console.log("Clone not selected campaign");
        w2alert(w2utils.lang('Select campaign for clone.'));
    } else {
        w2confirm(w2utils.lang('Are you sure you want to clone a campaign?'), function (btn) {
            if (btn == 'Yes') {
                var id, name;
                $.ajax({
                    type: "GET",
                    //async: false,
                    dataType: 'json',
                    data: {"request": JSON.stringify({"cmd": "clone", "id": campaignId})},
                    url: '/api/campaigns'
                }).done(function(data) {
                    if (data['status'] == 'error') {
                        w2alert(w2utils.lang(data["message"]));
                    } else {
                        id = data["recid"];
                        name = data["name"];
                        w2ui.campaign.add({recid: id, name: name}, true);
                        w2ui.campaign.editField(id, 1);
                    }
                });
            }
        });
    }
}

function addCampaign(groupId) {
    if (isNaN(groupId)) {
        w2alert(w2utils.lang('Select group.'));
    } else {
        var id, name;
        $.ajax({
            type: "GET",
            //async: false,
            dataType: 'json',
            data: {"request": JSON.stringify({"cmd": "add", "id": groupId})},
            url: '/api/campaigns'
        }).done(function(data) {
            if (data['status'] == 'error') {
                w2alert(w2utils.lang(data["message"]), w2utils.lang('Error'));
            } else {
                id = data["recid"];
                name = data["name"];
                w2ui.campaign.add({recid: id, name: name}, true);
                w2ui.campaign.editField(id, 1);
            }
        });
    }
}

// --- Get campaign data ---
function getCampaign(recid, name) {
    w2ui.layout.lock('main', w2utils.lang('Loading...'), true);
    $.ajax({
        type: "GET",
        url: '/api/campaign',
        dataType: 'json',
        data: {"request": JSON.stringify({"cmd": "get", "id": parseInt(recid)})}
    }).done(function(data) {
        zone = new Date().getTimezoneOffset() * 60;
        refreshProfilesList(data["profileId"]);
        refreshSenderList(data["senderId"]);
        $('#campaignId').val(recid);
        $('#campaignName').val(name);
        $("#campaignSubject").val(data["subject"]);
        $("#campaignStartDate").val(w2utils.formatDate((new Date((data["startDate"] + zone)* 1000 )), w2utils.settings.dateFormat));
        $("#campaignStartTime").val(w2utils.formatTime((new Date((data["startDate"] + zone) * 1000)), w2utils.settings.timeFormat));
        $("#campaignEndDate").val(w2utils.formatDate((new Date((data["endDate"] + zone) * 1000)), w2utils.settings.dateFormat));
        $("#campaignEndTime").val(w2utils.formatTime((new Date((data["endDate"] + zone) * 1000)), w2utils.settings.timeFormat));
        $("#campaignSendUnsubscribe").prop("checked", data["sendUnsubscribe"]);
        $("#campaignCompressHTML").prop("checked", data["compressHTML"]);
        $("#campaignTemplate").val(data["template"]);
        $('#campaignAcceptSend').prop('checked', data["accepted"]);

        cm.setValue(data["template"]);

        w2ui['recipient'].postData["campaign"] = parseInt(recid);
        w2ui.layout.unlock('main');
        w2ui['sidebar'].click('parameter');
    });
}

// ---Save campaign data ---
function saveCampaign() {
    if ($('#campaignAcceptSend').is(':checked')) {
        w2alert(w2utils.lang("You can't save an accepted for send campaign."), w2utils.lang('Error'));
    } else {
        w2confirm(w2utils.lang('Save changes in campaign?'), function (btn) {
            if ( btn == 'Yes') {
                // ---Save campaign data ---
                w2ui.layout.lock('main', w2utils.lang('Saving...'), true);
                $.ajax({
                    type: "POST",
                    url: '/api/campaign',
                    dataType: "json",
                    data: {"request": JSON.stringify(
                            {
                                "cmd": "save",
                                "id": parseInt($('#campaignId').val()),
                                "content": {
                                    "profileId": $('#campaignProfileId').data('selected').id,
                                    "name": $('#campaignName').val(),
                                    "subject": $("#campaignSubject").val(),
                                    "senderId": $('#campaignSenderId').data('selected').id,
                                    "startDate": getDate($("#campaignStartDate").val(), $("#campaignStartTime").val()),
                                    "endDate": getDate($("#campaignEndDate").val(), $("#campaignEndTime").val()),
                                    "compressHTML": $("#campaignCompressHTML").is(":checked"),
                                    "sendUnsubscribe": $("#campaignSendUnsubscribe").is(":checked"),
                                    "template": cm.getValue()
                                }
                            }
                        )},
                    dataType: "json"
                }).done(function(data) {
                    if (data['status'] == 'error') {
                        w2alert(w2utils.lang(data["message"]), w2utils.lang('Error'));
                    } else {
                        w2ui.layout.lock('main', w2utils.lang('Saved'), false);
                        w2ui['sidebar'].click('parameter');
                    }
                    setTimeout(function(){
                        w2ui.layout.unlock('main');
                    }, 1500);
                });
            }
        });
    }
}

function refreshProfilesList(selectedProfile){
    var profile;
    $.ajax({
        type: "GET",
        async: false,
        url: '/api/profilelist',
        data: {"request": JSON.stringify({"cmd": "get"})},
        dataType: "json"
    }).done(function(data) {
        profile = data;
    });
    w2ui['parameter'].set('campaignProfileId', { options: { items: profile } });
    w2ui['parameter'].record['campaignProfileId'] = selectedProfile;
    w2ui['parameter'].refresh();
}

function refreshSenderList(selectedSender){
    var sender;
    $.ajax({
        type: "GET",
        async: false,
        url: '/api/senderlist',
        dataType: "json",
        data: {"request": JSON.stringify({"cmd": "get", "id": parseInt(w2ui['group'].getSelection()[0])})},
    }).done(function(data) {
        sender = data;
    });
    w2ui['parameter'].set('campaignSenderId', { options: { items: sender } });
    w2ui['parameter'].record['campaignSenderId'] = selectedSender;
    w2ui['parameter'].refresh();
}
