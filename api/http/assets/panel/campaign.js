// --- Campaign table ---
$('#campaign').w2grid({
    header: w2utils.lang('Campaign'),
    show: {
        header: true,
        toolbar: true,
        footer: true,
        toolbarDelete: false,
        toolbarAdd: true,
        toolbarSave: true,
        toolbarSearch: false
    },
    name: 'campaign',
    columns: [
        { field: 'recid', caption: w2utils.lang('Id'), size: '50px', style: 'background-color: #efefef; border-bottom: 1px solid white; padding-right: 5px;', attr: "align=right" },
        { field: 'name', caption: w2utils.lang('Name'), size: '100%', editable: { type: 'text' } }
    ],
    multiSelect: false,
    sortData: [{ field: 'recid', direction: 'DESC' }],
    url: '/api/campaigns',
    method: 'GET',
    onAdd: function (event) {
        var id, name;
        $.ajax({
            type: "GET",
            dataType: 'json',
            data: {"request": JSON.stringify({"cmd": "add", "id": parseInt(w2ui['group'].getSelection()[0])})},
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

    },
    onSelect: function (event) {
        w2ui.layout.lock('main', w2utils.lang('Loading...'), true);
        var record = this.get(event.recid);
        // --- Get campaign data ---
        $.ajax({
            type: "GET",
            url: '/api/campaign',
            dataType: 'json',
            data: {"request": JSON.stringify({"cmd": "get", "id": parseInt(event.recid)})}
        }).done(function(data) {
            zone = new Date().getTimezoneOffset() * 60;
            refreshProfilesList(data["profileId"]);
            refreshSenderList(data["senderId"]);
            $('#campaignId').val(record.recid);
            $('#campaignName').val(record.name);
            $("#campaignSubject").val(data["subject"]);
            $("#campaignStartDate").val(w2utils.formatDate((new Date((data["startDate"] + zone)* 1000 )), w2utils.settings.dateFormat));
            $("#campaignStartTime").val(w2utils.formatTime((new Date((data["startDate"] + zone) * 1000)), w2utils.settings.timeFormat));
            $("#campaignEndDate").val(w2utils.formatDate((new Date((data["endDate"] + zone) * 1000)), w2utils.settings.dateFormat));
            $("#campaignEndTime").val(w2utils.formatTime((new Date((data["endDate"] + zone) * 1000)), w2utils.settings.timeFormat));
            $("#campaignSendUnsubscribe").prop("checked", data["sendUnsubscribe"]);
            $("#campaignTemplate").val(data["template"]);
            $('#campaignAcceptSend').prop('checked', data["accepted"]);

            CKEDITOR.instances.campaignTemplate.setData(data["template"]);

            w2ui['recipient'].postData["campaign"] = parseInt(event.recid)
            w2ui.layout.unlock('main');
            w2ui['sidebar'].click('parameter');
        });
        // --- /Get campaign data ---
    }
});
// --- /Campaign table ---