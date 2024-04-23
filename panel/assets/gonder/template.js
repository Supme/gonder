var cmHTML = CodeMirror.fromTextArea(document.getElementById("campaignTemplateHTML"), {
    lineNumbers: true,
    mode: {
        name: "htmlmixed",
        scriptTypes: [{matches: /\/x-handlebars-template|\/x-mustache/i,mode: null}]
    },
    selectionPointer: true,
    theme: "dracula"
});

var cmAMP = CodeMirror.fromTextArea(document.getElementById("campaignTemplateAMP"), {
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
        { id: 'preview', text: w2utils.lang('Preview') },
        { id: 'html', text: w2utils.lang('HTML') },
        { id: 'text', text: w2utils.lang('Text') },
        { id: 'amp', text: w2utils.lang('AMP') },
        { id: 'help', text: w2utils.lang('Help') }
    ],
    onClick: function (event) {
        switch (event.target)
        {
            case "preview":
                templateShowPreview();
                break;
            case "html":
                templateShowHTML();
                break;
            case "text":
                templateShowText();
                break;
            case "amp":
                templateShowAMP();
                break;
            case "help":
                templateShowHelp();
                break;
        }
    }
});


function templateShowAMP() {
    $("#campaignTemplatePreviewContainer").hide();
    $("#campaignTemplateHelpContainer").hide();
    $("#campaignTemplateHTMLContainer").hide();
    $("#campaignTemplateTextContainer").hide();
    $("#campaignTemplateAMPContainer").show();
    cmHTML.refresh();
    cmAMP.refresh();
}

function templateShowText() {
    $("#campaignTemplatePreviewContainer").hide();
    $("#campaignTemplateHelpContainer").hide();
    $("#campaignTemplateHTMLContainer").hide();
    $("#campaignTemplateAMPContainer").hide();
    $("#campaignTemplateTextContainer").show();
    cmHTML.refresh();
    cmAMP.refresh();
}

function templateShowHTML() {
    $("#campaignTemplatePreviewContainer").hide();
    $("#campaignTemplateHelpContainer").hide();
    $("#campaignTemplateTextContainer").hide();
    $("#campaignTemplateAMPContainer").hide();
    $("#campaignTemplateHTMLContainer").show();
    cmHTML.refresh();
    cmAMP.refresh();
}

function templateShowPreview() {
    $("#campaignTemplateHTMLContainer").hide();
    $("#campaignTemplateHelpContainer").hide();
    $("#campaignTemplateTextContainer").hide();
    $("#campaignTemplateAMPContainer").hide();
    document.getElementById("campaignTemplatePreview").srcdoc = cmHTML.getValue();
    $("#campaignTemplatePreviewContainer").show();
}

function templateShowHelp() {
    $("#campaignTemplatePreviewContainer").hide();
    $("#campaignTemplateHTMLContainer").hide();
    $("#campaignTemplateTextContainer").hide();
    $("#campaignTemplateAMPContainer").hide();
    $("#campaignTemplateHelpContainer").show();
}

function MakeTextFromHTML(withImg) {
    var config;
    if (withImg) {
        config = {
            headingStyle: "hashify",
            linkProcess: function (href, linkText) {
                href = href.replace(/^\s*?(\[.*?\]).*?/g, '');
                if (linkText == "") {
                    return "(" + href + ")";
                }
                return "[" + linkText + "] " + "(" + href + ")";
            }
        };
    } else {
        config = {
            headingStyle: "hashify",
            imgProcess: function (src, alt){
                if (alt == "") {
                    return " ";
                }
                return alt
            },
            linkProcess: function (href, linkText) {
                href = href.replace(/^\s*?(\[.*?\]).*?/g, '');
                if (linkText == " ") {
                    return "(" + href + ")";
                }
                return "[" + linkText + "] " + "(" + href + ")";
            }
        };
    }

    $("#campaignTemplateText").val(
        htmlToPlainText(cmHTML.getValue().replace(/(?=<!--)([\s\S]*?)-->/g, ''), config).replace(/(&\S{2,16};)/g, function(str, num) {
            return $("<span />", { html: num }).text();
        })
    );
}
