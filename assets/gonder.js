tinymce.init({
                selector:'textarea#message',
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
                    "textcolor","advlist autolink lists link image charmap print preview anchor",
                    "searchreplace code fullscreen",
                    'codemirror',
                    "insertdatetime media table contextmenu paste"
                ],
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
                save_enablewhendirty: true,
                theme: "modern",
                height : 600,
                width : 800,
                file_browser_callback : fm
            });


function fm(field_name, url, type, win) {
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
            width: 850,
            height: 650,
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