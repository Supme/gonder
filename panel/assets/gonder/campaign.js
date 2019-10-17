// --- Campaign table ---
w2ui['bottom'].content('main', $().w2grid({
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
            {type: 'break'},
            {id: 'clone', type: 'button', caption: w2utils.lang('Clone'), icon: 'w2ui-icon-columns'},
            {type: 'break'},
            {id: 'reports', type: 'menu-radio', icon: 'w2ui-icon-info', items: [
                    { id: 'recipients', text: w2utils.lang('Recipients')},
                    { id: 'clicks', text: w2utils.lang('Clicks')},
                    { id: 'unsubscribed', text: w2utils.lang('Unsubscribed')},
                    { id: 'question', text: w2utils.lang('Question')}
                ],
                text: function (item) {
                    var el   = this.get('reports:' + item.selected);
                    return w2utils.lang('Report: ') + el.text;
                },
                selected: 'recipients'
            },
            {id: 'download', type: 'button', caption: w2utils.lang('Download')}
        ],

        onClick: function (event) {
            if (event.target === 'download') {
                var campaignId = w2ui.campaign.getSelection();
                if (campaignId.length === 0) {
                    w2alert(w2utils.lang('Select campaign for download this report.'));
                    return;
                }
                loadLink('/report/campaign?id='+ w2ui.campaign.getSelection()[0] + '&type=' + this.get('reports').selected +'&format=csv');
                return
            }

            if (event.target === 'clone')
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
}));
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
    var campaignData = getCampaignData(recid);

    refreshProfilesList(campaignData.profiles, campaignData.profileID);
    refreshSenderList(campaignData.senders, campaignData.senderID);

    $('#campaignId').val(recid);
    $('#campaignName').val(name);
    $("#campaignSubject").val(campaignData.subject);
    $("#campaignStartDate").val(w2utils.formatDate(campaignData.startDate, w2utils.settings.dateFormat));
    $("#campaignStartTime").val(w2utils.formatTime(campaignData.startDate, w2utils.settings.timeFormat));
    $("#campaignEndDate").val(w2utils.formatDate(campaignData.endDate, w2utils.settings.dateFormat));
    $("#campaignEndTime").val(w2utils.formatTime(campaignData.endDate, w2utils.settings.timeFormat));
    $("#campaignSendUnsubscribe").prop("checked", campaignData.sendUnsubscribe);
    $("#campaignCompressHTML").prop("checked", campaignData.compressHTML);
    $("#campaignTemplateHTML").val(campaignData.templateHTML);
    $("#campaignTemplateText").val(campaignData.templateText);

    setAcceptSend(campaignData.accepted);

    cm.setValue(campaignData.templateHTML);

    w2ui['recipient'].postData["campaign"] = parseInt(recid);
    w2ui.layout.unlock('main');

    w2ui['toolbar'].click('parametersButton');
}

function getCampaignData(campaignId) {
    var campaignData = {
        subject: "",
        profiles: [{}],
        profileID: 0,
        profileName: "",
        senders: [{}],
        senderID: 0,
        senderName: "",
        startDate: Date(),
        endDate: Date(),
        sendUnsubscribe: false,
        compressHTML: false,
        accepted: false,
        templateHTML: "",
        templateText: ""
    };
    $.ajax({
        type: "GET",
        async: false,
        url: '/api/campaign',
        dataType: 'json',
        data: {"request": JSON.stringify({"cmd": "get", "id": parseInt(campaignId)})}
    }).done(function(data) {
        campaignData.subject = data["subject"];
        campaignData.profileID = data["profileId"];
        campaignData.senderID = data["senderId"];
        // time from server in UTC, add offset
        var zone = new Date().getTimezoneOffset() * 60000;
        campaignData.startDate = new Date((data["startDate"] * 1000) + zone);
        campaignData.endDate = new Date((data["endDate"] * 1000) + zone);
        campaignData.sendUnsubscribe = data["sendUnsubscribe"];
        campaignData.compressHTML = data["compressHTML"];
        campaignData.accepted = data["accepted"];
        campaignData.templateHTML = data["templateHTML"];
        campaignData.templateText = data["templateText"];
    });
    $.ajax({
        type: "GET",
        async: false,
        url: '/api/profilelist',
        data: {"request": JSON.stringify({"cmd": "get"})},
        dataType: "json"
    }).done(function(data) {
        campaignData.profiles = data;
        campaignData.profiles.forEach(function(v) {
           if (v.id === campaignData.profileID) {
               campaignData.profileName = v.text;
           }
        });
    });
    $.ajax({
        type: "GET",
        async: false,
        url: '/api/senderlist',
        dataType: "json",
        data: {"request": JSON.stringify({"cmd": "get", "id": parseInt(w2ui['group'].getSelection()[0])})},
    }).done(function(data) {
        campaignData.senders = data;
        campaignData.senders.forEach(function(v) {
            if (v.id === campaignData.senderID) {
                campaignData.senderName = v.text;
            }
        });
    });

    return campaignData;
}

function getTimestamp(dateStr, timeStr) {
    var d = new Date(w2utils.isDateTime(dateStr + ' ' + timeStr, w2utils.settings.datetimeFormat, true));
    return (d.getTime() - (d.getTimezoneOffset() * 60000))/1000;
}

// ---Save campaign data ---
function saveCampaign() {
    if (w2ui['toolbar'].get('acceptSend').checked) {
        w2alert(w2utils.lang("You can't save an accepted for send campaign."), w2utils.lang('Error'));
        return
    }
    w2confirm(w2utils.lang('Save changes in campaign?'), function (btn) {
        if (btn === 'Yes') {
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
                                "startDate": getTimestamp($("#campaignStartDate").val(), $("#campaignStartTime").val()),
                                "endDate": getTimestamp($("#campaignEndDate").val(), $("#campaignEndTime").val()),
                                "compressHTML": $("#campaignCompressHTML").is(":checked"),
                                "sendUnsubscribe": $("#campaignSendUnsubscribe").is(":checked"),
                                "templateHTML": cm.getValue(),
                                "templateText": $("#campaignTemplateText").val()
                            }
                        }
                    )}
            }).done(function(data) {
                if (data['status'] === 'error') {
                    w2alert(w2utils.lang(data["message"]), w2utils.lang('Error'));
                } else {
                    w2ui.layout.lock('main', w2utils.lang('Saved'), false);
                }
                setTimeout(function(){
                    w2ui.layout.unlock('main');
                }, 1000);
            });
        }
    });
}

function refreshProfilesList(profiles, profileID){
    w2ui['parameter'].set('campaignProfileId', { options: { items: profiles } });
    w2ui['parameter'].record['campaignProfileId'] = profileID;
    w2ui['parameter'].refresh();
}

function refreshSenderList(senders, senderID){
    w2ui['parameter'].set('campaignSenderId', { options: { items: senders } });
    w2ui['parameter'].record['campaignSenderId'] = senderID;
    w2ui['parameter'].refresh();
}
