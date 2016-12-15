function getDate(dateStr, timeStr) {
    //ToDo parse config format
    //w2utils.settings.dateFormat
    var dateParts = dateStr.split("/");
    var timeParts = timeStr.split(":");
    return new Date(dateParts[2],(dateParts[1] - 1),dateParts[0],timeParts[0],timeParts[1],0);
}

// --- Config for layout ---
var config = {
    sidebar: {
        name: 'sidebar',
        topHTML: '<div style="padding: 10px 5px; border-bottom: 1px solid #99bbe8";><span style="text-transform: uppercase;">'+w2utils.lang("Menu")+'</span></div>',
        nodes: [
            {
                id: 'campaign', text: w2utils.lang('Campaign'), group: true, expanded: true, nodes:
                [
                    {id: 'parameter', text: w2utils.lang('Parameters'), img: 'icon-page'},
                    {id: 'editor', text: w2utils.lang('Editor'), img: 'w2ui-icon-pencil'},
                    {id: 'recipient', text: w2utils.lang('Recipients'), img: 'w2ui-icon-columns'},
                    {id: 'save', text: w2utils.lang('Save'), img: 'w2ui-icon-check'}
                ]
            },
            {
                id: 'settings', text: w2utils.lang('Settings'), group: true, expanded: false, nodes:
                [
                    {id: 'status', text: w2utils.lang('Status'), img: 'w2ui-icon-info'},
                    {id: 'users', text: w2utils.lang('Users'), img: 'w2ui-icon-columns'},
                    {id: 'profile', text: w2utils.lang('Profiles'), img: 'icon-page'}
                ]
            }
        ],
        onClick: function (event) {
            if ($('#campaignId').val() != '' || event.target =='profile'  || event.target =='status' || event.target =='users') {
                switch (event.target) {
                    case 'parameter':
                        $('#template').hide();
                        $('#recipient').hide();
                        $('#parameter').show();
                        $('#profile').hide();
                        $('#users').hide();
                        $('#status').hide();
                        break;
                    case 'editor':
                        $('#parameter').hide();
                        $('#recipient').hide();
                        $('#template').show();
                        $('#profile').hide();
                        $('#users').hide();
                        $('#status').hide();
                        break;
                    case 'recipient':
                        w2ui['recipient'].url = '/api/recipients';
                        w2ui['recipient'].reload();
                        $('#template').hide();
                        $('#parameter').hide();
                        $('#recipient').show();
                        $('#profile').hide();
                        $('#users').hide();
                        $('#status').hide();
                        break;
                    case 'save':
                        w2confirm(w2utils.lang('Save changes in campaign?'), function (btn) {
                            if ( btn == 'Yes') {
                                // ---Save campaign data ---
                                w2ui.layout.lock('main', w2utils.lang('Saving...'), true);

                                $.ajax({
                                    type: "POST",
                                    url: '/api/campaign',
                                    data: {"request": JSON.stringify(
                                        {
                                            "cmd": "save",
                                            "id": parseInt($('#campaignId').val()),
                                            "content": {
                                                "profileId": $('#campaignProfileId').data('selected').id,
                                                "name": $('#campaignName').val(),
                                                "subject": $("#campaignSubject").val(),
                                                "senderId": $('#campaignSenderId').data('selected').id,
                                                "startDate": getDate($("#campaignStartDate").val(), $("#campaignStartTime").val()).getTime() / 1000,
                                                "endDate": getDate($("#campaignEndDate").val(), $("#campaignEndTime").val()).getTime() / 1000,
                                                "sendUnsubscribe": $("#campaignSendUnsubscribe").is(":checked"),
                                                "template": CKEDITOR.instances.campaignTemplate.getData() == '' ? $("#campaignTemplate").val() : CKEDITOR.instances.campaignTemplate.getData()
                                            }
                                        }
                                    )},
                                    dataType: "json"
                                }).done(function(data) {
                                    if (data['status'] == 'error') {
                                        w2alert(w2utils.lang(data["message"]), w2utils.lang('Error'));
                                    }
                                    w2ui.layout.lock('main', w2utils.lang('Saved'), false);
                                    w2ui['sidebar'].click('parameter');
                                    setTimeout(function(){
                                        w2ui.layout.unlock('main');
                                    }, 1500);
                                });
                            }
                        });
                        // ToDo this
                        w2ui['sidebar'].select(w2ui['sidebar'].selected);
                        w2ui['sidebar'].unselect('save');
                        break;

                    case 'profile':
                        w2ui['profile'].url = '/api/profiles';
                        w2ui['profile'].reload();
                        $('#template').hide();
                        $('#parameter').hide();
                        $('#recipient').hide();
                        $('#profile').show();
                        $('#users').hide();
                        $('#status').hide();
                        break;
                    case 'users':
                        $('#template').hide();
                        $('#parameter').hide();
                        $('#recipient').hide();
                        $('#profile').hide();
                        $('#users').show();
                        $('#status').hide();
                        break;
                    case 'status':
                        $('#template').hide();
                        $('#parameter').hide();
                        $('#recipient').hide();
                        $('#profile').hide();
                        $('#users').hide();
                        $('#status').show();
                        break;
                }
                w2ui['layout'].resize();
            } else {
                w2ui['sidebar'].unselect(event.target);
                w2alert(w2utils.lang('Select group and campaign, before select action.'));
            }
        }
    }
};
// --- /Config for layout ---

// --- Layout ---
var pstyle = 'border: 1px solid #dfdfdf; padding: 2px;';
$('#layout').w2layout({
    name: 'layout',
    panels: [
        { type: 'top', size:32, style: pstyle, content: "<div style='text-align: center;'><img src='/assets/img/logo.png' height='20px' border='0px'/><span style='font-size: 20px;'> Mass email sender</span></div>" },
        { type: 'left', size: 200, resizable: true, style: pstyle },
        { type: 'main', hidden: true, style: pstyle},
        { type: 'bottom', size: 250, resizable: true, style: pstyle }
    ]
});
w2ui.layout.content('left', $().w2sidebar(config.sidebar));
w2ui.layout.content('main', $('#formbox').html());
w2ui.layout.content('bottom', $('#bottom').html());

// --- Parameters form ---
$('#parameter').w2form({
    header: w2utils.lang("Parameters"),
    name: 'parameter',
    fields: [
        { name: 'campaignId', type: 'text', html: { caption: w2utils.lang('Id'), attr: 'size="4" readonly' } },
        { name: 'campaignName', type: 'text', html: { caption: w2utils.lang('Name'), attr: 'size="40" readonly' } },
        { name: 'campaignProfileId', type: 'list', html: { caption: w2utils.lang('Profile'), attr: 'size="40"' }, minLength: 0},
        { name: 'campaignSubject', type: 'text', html: { caption: w2utils.lang('Subject'), attr: 'size="40"' } },
        { name: 'campaignSenderId', type: 'list', html: { caption: w2utils.lang('Sender'), attr: 'size="40"' }, minLength: 0},
        { name: 'campaignStartDate', type: 'date', html: { caption: w2utils.lang('Start date'), attr: 'size="10"' } },
        { name: 'campaignStartTime', type: 'time', html: { caption: w2utils.lang('Start time'), attr: 'size="10"' } },
        { name: 'campaignEndDate', type: 'date', html: { caption: w2utils.lang('End date'), attr: 'size="10"' } },
        { name: 'campaignEndTime', type: 'time', html: { caption: w2utils.lang('End time'), attr: 'size="10"' } },
        { name: 'campaignSendUnsubscribe', type: 'checkbox', html: { caption: w2utils.lang('Send unsubscribe') } },
        { name: 'campaignAcceptSend', type: 'toggle', html: { caption: w2utils.lang('Accept send') } }
    ]
});

$('#campaignSendUnsubscribe').click(function(data) {
    if ($('#campaignSendUnsubscribe').is(':checked')) {
        w2confirm(w2utils.lang('You are sure send mail for unsubscribed?'), function (btn) {
            if (btn != 'Yes') {
                $('#campaignSendUnsubscribe').prop('checked', !$('#campaignSendUnsubscribe').is(':checked'));
            }
        });
    }
});

$('#campaignAcceptSend').click(function(data) {
    var confirm;
    if ($('#campaignAcceptSend').is(':checked')) {
        confirm = w2utils.lang('You are sure to activate campaign?');
    } else {
        confirm = w2utils.lang('You are sure to deactivate campaign?');
    }
    w2confirm(w2utils.lang(confirm), function (btn) {
        if (btn == 'Yes') {
            $.ajax({
                type: "POST",
                url: '/api/campaign',
                dataType: "json",
                data: {"request": JSON.stringify({"cmd": "accept", "id": parseInt($('#campaignId').val()), "select": $('#campaignAcceptSend').is(':checked')})}
            }).done(function (data) {
                if (data['status'] == 'error') {
                    w2alert(w2utils.lang(data["message"]), w2utils.lang('Error'));
                    $('#campaignAcceptSend').prop('checked', !$('#campaignAcceptSend').is(':checked'));
                }
            })
        } else {
            $('#campaignAcceptSend').prop('checked', !$('#campaignAcceptSend').is(':checked'));

        }
    });
});

// --- /Parameters form ---

// --- Init ---
$('#template').hide();
$('#recipient').hide();
$('#parameter').hide();
$('#profile').hide();
$('#status').hide();
$('#users').hide();
// --- /Init ---

// --- /Layout ---
