function getDate(dateStr, timeStr) {
    var d = new Date(w2utils.isDateTime(dateStr + ' ' + timeStr, w2utils.settings.datetimeFormat, true));
    return d.getTime()/1000;
}

// --- Layout ---
$('#layout').w2layout({
    name: 'layout',
    panels: [
        { type: 'top', size: 35, style: "padding: 4px; border: 1px solid #dfdfdf; border-radius: 3px;" },
        { type: 'main', hidden: true},
        { type: 'bottom', size: 250, resizable: true }
    ]
});

w2ui.layout.content('top', $().w2toolbar({
    name: 'toolbar',
    right : 'v' + version,
    items: [
        { type: 'radio', id: 'parametersButton', group: '1', text: w2utils.lang('Parameters'), img: 'icon-page' },
        { type: 'radio', id: 'editorButton', group: '1', text: w2utils.lang('Editor'), img: 'w2ui-icon-pencil' },
        { type: 'radio', id: 'recipientsButton', group: '1', text: w2utils.lang('Recipients'), img: 'w2ui-icon-columns' },
        { type: 'break' },
        { type: 'button', id: 'saveButton', text: w2utils.lang('Save'), img: 'w2ui-icon-check'},
        { type: 'break' },
        { type: 'check', id: 'acceptSend', group: '1', text: w2utils.lang('Accept send'), style: '.checked {background: #ddff00}' },
        { type: 'spacer' },
        { type: 'radio', id: 'statusButton', group: '1', text: w2utils.lang('Status'), img: 'w2ui-icon-info' },
        { type: 'radio', id: 'usersButton', group: '1', text: w2utils.lang('Users'), img: 'w2ui-icon-columns' },
        { type: 'radio', id: 'profilesButton', group: '1', text: w2utils.lang('Profiles'), img: 'icon-page' },
        { type: 'break' }
    ],
    onClick: function (event) {
        // console.log('Target: '+ event.target, event);
        if (
            $('#campaignId').val() != '' ||
            event.target =='profilesButton'  ||
            event.target =='statusButton' ||
            event.target =='usersButton'
        ) {
            switch (event.target) {
                case 'parametersButton':
                    switchToParameters();
                    break;
                case 'editorButton':
                    switchToEditor();
                    break;
                case 'recipientsButton':
                    switchToRecipients();
                    break;

                case 'saveButton':
                    saveCampaign();
                    break;

                case 'acceptSend':
                    switchAcceptSend();
                    break;

                case 'profilesButton':
                    switchToProfiles();
                    break;
                case 'usersButton':
                    switchToUsers();
                    break;
                case 'statusButton':
                    switchToStatus();
                    break;

            }
            w2ui['layout'].resize();
        } else {
            console.log(event.target);
            event.checked = false;
            w2alert(w2utils.lang('Select group and campaign, before select action.'));
        }
    }
}));

w2ui.layout.content('main', $('#formbox').html());
w2ui.layout.content('bottom', $().w2layout({
    name: 'bottom',
    panels: [
        { type: 'left', size: '50%', resizable: true },
        { type: 'main',  size: '50%', resizable: true }
    ]
}));

function switchAcceptSend() {
    var confirm;
    if (w2ui['toolbar'].get('acceptSend').checked) {
        confirm = w2utils.lang('Are you sure to deactivate campaign?');
    } else {
        confirm = 'Are you sure to activate campaign?';
    }
    w2confirm(w2utils.lang(confirm), function (btn) {
        if (btn == 'Yes') {
            $.ajax({
                type: "POST",
                url: '/api/campaign',
                dataType: "json",
                data: {"request": JSON.stringify({"cmd": "accept", "id": parseInt($('#campaignId').val()), "select": w2ui['toolbar'].get('acceptSend').checked})}
            }).done(function (data) {
                if (data['status'] == 'error') {
                    w2alert(w2utils.lang(data["message"]), w2utils.lang('Error'));
                    setAcceptSend(!w2ui['toolbar'].get('acceptSend').checked);
                }
            })
        } else {
            setAcceptSend(!w2ui['toolbar'].get('acceptSend').checked);
        }
    });
}

function setAcceptSend(accept) {
    if (accept) {
        w2ui['toolbar'].check('acceptSend');
    } else {
        w2ui['toolbar'].uncheck('acceptSend');
    }
}

function switchToParameters() {
    $('#template').hide();
    $('#recipient').hide();
    $('#parameter').show();
    $('#profile').hide();
    $('#users').hide();
    $('#status').hide();
}

function switchToEditor() {
    $('#parameter').hide();
    $('#recipient').hide();
    $('#template').show();
    w2ui.templateTabs.click('preview');
    $('#profile').hide();
    $('#users').hide();
    $('#status').hide();
}

function switchToRecipients() {
    w2ui['recipient'].url = '/api/recipients';
    w2ui['recipient'].reload();
    $('#template').hide();
    $('#parameter').hide();
    $('#recipient').show();
    $('#profile').hide();
    $('#users').hide();
    $('#status').hide();
}

function switchToProfiles() {
    w2ui['profile'].url = '/api/profiles';
    w2ui['profile'].reload();
    $('#template').hide();
    $('#parameter').hide();
    $('#recipient').hide();
    $('#profile').show();
    $('#users').hide();
    $('#status').hide();
}

function switchToUsers() {
    w2ui['userList'].url = '/api/users';
    w2ui['userList'].reload();
    w2ui['unitList'].url = '/api/units';
    w2ui['unitList'].reload();
    $('#template').hide();
    $('#parameter').hide();
    $('#recipient').hide();
    $('#profile').hide();
    $('#users').show();
    $('#status').hide();
}

function switchToStatus() {
    $('#template').hide();
    $('#parameter').hide();
    $('#recipient').hide();
    $('#profile').hide();
    $('#users').hide();
    $('#status').show();
}

// --- /Layout ---

// --- Parameters form ---
$('#parameter').w2form({
    name: 'parameter',
    fields: [
        { name: 'campaignId', type: 'text', html: { caption: w2utils.lang('Id'), attr: 'size="4" readonly' } },
        { name: 'campaignName', type: 'text', html: { caption: w2utils.lang('Name'), attr: 'size="40" readonly' } },
        { name: 'campaignProfileId', type: 'list', html: { caption: w2utils.lang('Profile'), attr: 'size="40"' }, minLength: 0},
        { name: 'campaignSubject', type: 'text', html: { caption: w2utils.lang('Subject'), attr: 'size="40" autocomplete="off"' } },
        { name: 'campaignSenderId', type: 'list', html: { caption: w2utils.lang('Sender'), attr: 'size="40"' }, minLength: 0},
        { name: 'campaignStartDate', type: 'date', html: { caption: w2utils.lang('Start date'), attr: 'size="10" autocomplete="off"' }, options: {format: w2utils.settings.dateFormat} },
        { name: 'campaignStartTime', type: 'time', html: { caption: w2utils.lang('Start time'), attr: 'size="10" autocomplete="off"'}, options: {format: w2utils.settings.timeFormat} },
        { name: 'campaignEndDate', type: 'date', html: { caption: w2utils.lang('End date'), attr: 'size="10" autocomplete="off"' }, options: {format: w2utils.settings.dateFormat} },
        { name: 'campaignEndTime', type: 'time', html: { caption: w2utils.lang('End time'), attr: 'size="10" autocomplete="off"' }, options: {format: w2utils.settings.timeFormat} },
        { name: 'campaignCompressHTML', type: 'checkbox', html: { caption: w2utils.lang('Compress HTML') } },
        { name: 'campaignSendUnsubscribe', type: 'checkbox', html: { caption: w2utils.lang('Send unsubscribe') } },
        // { name: 'campaignAcceptSend', type: 'toggle', html: { caption: w2utils.lang('Accept send') } }
    ]
});

$('#campaignSendUnsubscribe').click(function(data) {
    if ($('#campaignSendUnsubscribe').is(':checked')) {
        w2confirm(w2utils.lang('Are you sure send mail for unsubscribed?'), function (btn) {
            if (btn != 'Yes') {
                $('#campaignSendUnsubscribe').prop('checked', !$('#campaignSendUnsubscribe').is(':checked'));
            }
        });
    }
});

// $('#campaignAcceptSend').click(function() {
//     var confirm;
//     if ($('#campaignAcceptSend').is(':checked')) {
//         confirm = 'Are you sure to activate campaign?';
//     } else {
//         confirm = w2utils.lang('Are you sure to deactivate campaign?');
//     }
//     w2confirm(w2utils.lang(confirm), function (btn) {
//         if (btn == 'Yes') {
//             $.ajax({
//                 type: "POST",
//                 url: '/api/campaign',
//                 dataType: "json",
//                 data: {"request": JSON.stringify({"cmd": "accept", "id": parseInt($('#campaignId').val()), "select": $('#campaignAcceptSend').is(':checked')})}
//             }).done(function (data) {
//                 if (data['status'] == 'error') {
//                     w2alert(w2utils.lang(data["message"]), w2utils.lang('Error'));
//                     $('#campaignAcceptSend').prop('checked', !$('#campaignAcceptSend').is(':checked'));
//                 }
//             })
//         } else {
//             $('#campaignAcceptSend').prop('checked', !$('#campaignAcceptSend').is(':checked'));
//
//         }
//     });
// });

// --- /Parameters form ---

