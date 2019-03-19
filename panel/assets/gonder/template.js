var cm = CodeMirror.fromTextArea(document.getElementById("campaignTemplateCode"), {
    lineNumbers: true,
    mode: {
        name: "htmlmixed",
        scriptTypes: [{matches: /\/x-handlebars-template|\/x-mustache/i,mode: null}]
    },
    selectionPointer: true,
    theme: "dracula"
});


$('#templateTabs').w2tabs({
    name: 'templateTabs',
    active: 'preview',
    tabs: [
        { id: 'preview', caption: w2utils.lang('Preview') },
        { id: 'code', caption: w2utils.lang('Code') },
        { id: 'help', caption: w2utils.lang('Help') },
    ],
    onClick: function (event) {
        switch (event.target)
        {
            case "preview":
                templateShowPreview();
                break;
            case "code":
                templateShowCode();
                break;
            case "help":
                templateShowHelp();
                break;
        }
    }
});

function templateShowCode() {
    $("#campaignTemplatePreviewContainer").hide();
    $("#campaignTemplateHelpContainer").hide();
    $("#campaignTemplateCodeContainer").show();
    cm.refresh();
}

function templateShowPreview() {
    $("#campaignTemplateCodeContainer").hide();
    $("#campaignTemplateHelpContainer").hide();
    $('#campaignTemplatePreview').html(cm.getValue());
    $("#campaignTemplatePreviewContainer").show();
}

function templateShowHelp() {
    $("#campaignTemplatePreviewContainer").hide();
    $("#campaignTemplateCodeContainer").hide();
    $("#campaignTemplateHelpContainer").show();
}
