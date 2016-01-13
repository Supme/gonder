window.onload = function() {

    CKEDITOR.replace( 'message', {
            filebrowserBrowseUrl: '/assets/filemanager/index.html',
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
            'docprops,'
        }
    );
};

//ToDo Косяк в ссылках с заменой & на &amp;
/*tinymce.init({
    selector:'textarea#message',
    force_br_newlines : false,
    force_p_newlines : false,
    forced_root_block : '',
    removed_menuitems: 'newdocument',
    toolbar:
            " core |" +
            " insertfile undo redo |" +
            " styleselect |" +
            " bold italic |" +
            " alignleft aligncenter alignright alignjustify |" +
            " bullist numlist |" +
            " link image |" +
            " forecolor backcolor |" +
            " textcolor |" +
            " code",
    plugins: [
        "textcolor","advlist autolink lists link image charmap preview anchor",
        "searchreplace code fullscreen",
        'codemirror',
        "insertdatetime table contextmenu paste",
        "autoresize", "autolink",
        "legacyoutput", "fullpage",
    ],

    fullpage_default_doctype: "<!DOCTYPE html>",
    fullpage_default_encoding: "UTF-8",

    codemirror: {
         indentOnInit: true, // Whether or not to indent code on init.
         path: 'codemirror-4.8', // Path to CodeMirror distribution
         config: {           // CodeMirror config object
              mode: 'application/x-httpd-php',
              lineNumbers: false
         },
              jsFiles: [          // Additional JS files to load
                   'mode/php/php.js',
                   'mode/htmlmixed/htmlmixed.js'
              ]
    },

    convert_urls : false,
    relative_urls: true,
    document_base_url: '',

    //entity_encoding : "raw",
    editor_encoding : "raw",
    //valid_elements : "*[*]",
    cleanup : false,
    cleanup_on_startup : false,
    //  save_enablewhendirty: true,

    paste_data_images: true,

    theme: "modern",
    height : 600,
    width : 800,

    file_browser_callback : function(field_name, url, type, win) {

        // from http://andylangton.co.uk/blog/development/get-viewport-size-width-and-height-javascript
        var w = window,
            d = document,
            e = d.documentElement,
            g = d.getElementsByTagName('body')[0],
            x = w.innerWidth || e.clientWidth || g.clientWidth,
            y = w.innerHeight|| e.clientHeight|| g.clientHeight;

        var cmsURL = '/assets/filemanager/index.html?&field_name='+field_name+'&langCode='+tinymce.settings.language;

        if(type == 'image') {
            cmsURL = cmsURL + "&type=images";
        }

        tinyMCE.activeEditor.windowManager.open({
            file : cmsURL,
            title : 'Filemanager',
            width : x * 0.8,
            height : y * 0.8,
            inline : "yes",
            resizable : "yes",
            close_previous : "no"
        });

    }
});
*/

/*

Dropzone.options.dropzone = {
    init: function() {

        var self = this;
        // config
        this.options.addRemoveLinks = true;
        this.options.dictRemoveFile = "Delete";
        this.options.clickable = false;

        thisDropzone = this;
        $.get('upload.php', function(data) {
            $.each(data, function (key, value) {
                var mockFile = {name: value.name, size: value.size};
                thisDropzone.options.addedfile.call(thisDropzone, mockFile);
                thisDropzone.options.thumbnail.call(thisDropzone, mockFile, "uploads/" + value.name);
            });
        });

        //New file added
        self.on("addedfile", function(file) {
            console.log('new file added ', file);
            file.previewTemplate.addEventListener("dblclick", function(e) {console.log("sfsfwfgwefg", e)})
        });

        self.on("success", function(file, responseText) {
            // Handle the responseText here. For example, add the text to the preview element:
        });

        // Send file starts
        self.on("sending", function(file) {
            console.log('upload started', file);
            $('.meter').show();
        });

        // File upload Progress
        self.on("totaluploadprogress", function(progress) {
            console.log("progress ", progress);
            $('.roller').width(progress + '%');
        });

        self.on("queuecomplete", function(progress) {
            console.log('queue complete', progress);
            $('.meter').delay(999).slideUp(999);
        });

        // On removing file
        self.on("removedfile", function(file) {
            console.log(file);
            $.ajax({
                url: './upload.php?file=' + file.name,
                type: 'DELETE',
                success: function(result) {
                    console.log(result);
                }
            });
        });

    }
};



/*
function fileman(field_name, url, type, win) {
        var roxyFileman = '/assets/fileman/';
        if (roxyFileman.indexOf("?") < 0) {
            roxyFileman += "?type=" + type;
        }
        else {
            roxyFileman += "&type=" + type;
        }
        roxyFileman += '&input=' + field_name + '&value=' + win.document.getElementById(field_name).value;
        if(tinyMCE.activeEditor.settings.language){
            roxyFileman += '&langCode=' + tinyMCE.activeEditor.settings.language;
        }
        tinyMCE.activeEditor.windowManager.open({
            file: roxyFileman,
            title: 'Roxy Fileman',
            width: 640,
            height: 480,
            resizable: "yes",
            plugins: "media",
            inline: "yes",
            close_previous: "no"
        },
        {
             window: win,
             input: field_name
        });
        return false;
    }
    */
