$("#campaignTemplateGrapesjsButton").html(
    '<button onclick="templateGrapesjsExport()">' + w2utils.lang("Export to code") + '</button>&nbsp;' +
    '<button onclick="templateGrapesjsImport();">' + w2utils.lang("Import from code") + '</button>&nbsp;' +
    '<button onclick="templateGrapesjsClear();">' + w2utils.lang("Clear") + '</button>&nbsp;'
);

$('#templateTabs').w2tabs({
    name: 'templateTabs',
    active: 'preview',
    tabs: [
        { id: 'preview', caption: w2utils.lang('Preview') },
        { id: 'code', caption: w2utils.lang('Code') },
        { id: 'help', caption: w2utils.lang('Help') },
        { id: 'grapesjs', caption: w2utils.lang('GrapesJs') }
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
            case "grapesjs":
                templateShowGrapesjs();
                break;
        }
    }
});

function templateShowCode() {
    $("#campaignTemplatePreviewContainer").hide();
    $("#campaignTemplateHelpContainer").hide();
    $("#campaignTemplateGrapesjsContainer").hide();
    $("#campaignTemplateCodeContainer").show();
    cm.refresh();
}

function templateShowPreview() {
    $("#campaignTemplateCodeContainer").hide();
    $("#campaignTemplateHelpContainer").hide();
    $("#campaignTemplateGrapesjsContainer").hide();
    $('#campaignTemplatePreview').html(cm.getValue());
    $("#campaignTemplatePreviewContainer").show();
}

function templateShowHelp() {
    $("#campaignTemplatePreviewContainer").hide();
    $("#campaignTemplateCodeContainer").hide();
    $("#campaignTemplateGrapesjsContainer").hide();
    $("#campaignTemplateHelpContainer").show();
}

function templateShowGrapesjs() {
    $("#campaignTemplatePreviewContainer").hide();
    $("#campaignTemplateCodeContainer").hide();
    $("#campaignTemplateHelpContainer").hide();
    $("#campaignTemplateGrapesjsContainer").show();
}

function templateGrapesjsImport() {
    var el = document.createElement( 'html' );
    el.innerHTML = cm.getValue();
    var body = el.getElementsByTagName('body');
    gjs.setComponents(body[0].innerHTML);
}

function templateGrapesjsExport() {
    var code = "" +
        "<!DOCTYPE html>\n" +
        "<html>\n" +
        "  <head>\n" +
        "    <meta http-equiv=\"Content-Type\" content=\"text/html; charset=utf-8\" />\n" +
        "    <meta name=\"viewport\" content=\"width=device-width; initial-scale=1.0; maximum-scale=1.0;\" />\n" +
        "    <title>" + $("#campaignSubject").val() + "</title>\n" +
        "    <style>\n" +
        "      " + gjs.getCss() + "\n" +
        "    </style>" +
        "  </head>\n" +
        "  <body bgcolor=\"#ffffff\" style=\"margin: 0; padding: 0;\">\n" +
        "    " + gjs.getHtml() + "\n" +
        "  </body>\n" +
        "</html>\n";
    cm.setValue(code);
}

function templateGrapesjsClear() {
    w2confirm(w2utils.lang('Clear GrapesJs template?'), function (btn) {
        if (btn == 'Yes') {
            gjs.setComponents("");
        }
    });
}
// cm.getValue()
// gjs.getHtml()
// cm.setValue(gjs.getHtml())
//