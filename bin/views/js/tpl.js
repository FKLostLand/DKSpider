var Html = function (info) {
    if (info.mode == client) {
        return logBoxHtml(client);
    }

    //先返回head
    var content = headHTML();

    content += '<body>\
    <div class="step2"> \
    <div id="a" class="split">\
        <form role="form" id="js-form" name="pholcus" onsubmit="return runStop();" method="POST" enctype="multipart/form-data">\
           <div id="c" class="split split-horizontal content">\
           <div class="col-md-12">\
             <!--<div class="box-header"><h3 class="box-title">All Spiders</h3></div>-->\
             <div class="box-body table-responsive no-padding" id="spider-box">\
               <table class="table table-hover">\
                 <tbody id="allSpiders">\
                   <tr>\
                     <th>#</th>\
                     <th>ID</th>\
                     <th>Name</th>\
                     <th>Description</th>\
                   </tr>' + spidersHtml(info.spiders) + '</tbody></table></div></div></div>\
            <div id="d" class="split split-horizontal content">\
            <div>\
              <div class="form-group">\
                <label>自定义配置（用“<>”包围，支持多关键字））</label>\
                <textarea name="Keyins" class="form-control" rows="2" placeholder="Enter ...">' + info.Keyins + '</textarea>\
              </div>\
            <div class="inline">\
              <div class="form-group">\
                <label>采集上限（默认限制URL数）</label>\
                <input name="Limit" type="number" class="form-control" min="0" value="' + info.Limit + '">\
              </div>' +
        ThreadNumHtml(info.ThreadNum) +
        PausetimeHtml(info.Pausetime) +
        ProxyMinuteHtml(info.ProxyMinute) +
        DockerCapHtml(info.DockerCap) +
        OutTypeHtml(info.OutType) +
        SuccessInheritHtml(info.SuccessInherit) +
        FailureInheritHtml(info.FailureInherit) +
        '</div>' +
        '</div></div>\
            <div class="box-footer">\
                ' + btnHtml(info.mode, info.status) +
        '</div>\
          </form>\
          </div>' + logBoxHtml(info.mode) + '</div>' + splitJSHTML() + '</body></html>';

    return content;
};

var spidersHtml = function (spiders) {
    var html = '';

    for (var i in spiders.menu) {
        html += '<tr>\
            <td>\
                <div class="checkbox">\
                  <label for="spider-' + i + '">\
                    <input name="spiders" id="spider-' + i + '" type="checkbox" value="' + spiders.menu[i].name + '"' +
            function () {
                if (spiders.curr[spiders.menu[i].name]) {
                    return "checked";
                }
                return
            }() + '>\
                  </label>\
                </div>\
            </td>\
            <td><label for="spider-' + i + '">' + i + '</label></td>\
            <td><label for="spider-' + i + '">' + spiders.menu[i].name + '</label></td>\
            <td><label for="spider-' + i + '">' + spiders.menu[i].description + '</label></td>\
        <tr>'
    }

    return html;
}
var ThreadNumHtml = function (ThreadNum) {
    return '<div class="form-group">\
                <label>并发协程</label>\
                <input name="ThreadNum" type="number" class="form-control" min="' + ThreadNum.min + '" max="' + ThreadNum.max + '" value="' + ThreadNum.curr + '">\
              </div>';
}

var DockerCapHtml = function (DockerCap) {
    return '<div class="form-group">\
                <label>分批输出限制</label>\
                <input name="DockerCap" type="number" class="form-control" min="' + DockerCap.min + '" max="' + DockerCap.max + '" value="' + DockerCap.curr + '">\
              </div>';
}

var PausetimeHtml = function (Pausetime) {
    var html = '<div class="form-group">\
                <label>暂停时长参考</label>\
                <select class="form-control" name="Pausetime">';
    for (var i in Pausetime.menu) {
        var isSelect = ""
        if (Pausetime.menu[i] == Pausetime.curr[0]) {
            isSelect = " selected";
        }
        ;
        if (Pausetime.menu[i] == 0) {
            html += '<option value="' + Pausetime.menu[i] + '"' + isSelect + '>' + "无暂停" + '</option>';
        } else {
            html += '<option value="' + Pausetime.menu[i] + '"' + isSelect + '>' + Pausetime.menu[i] + ' ms</option>';
        }
    }
    ;
    html += '</select></div>';
    return html;
}

var ProxyMinuteHtml = function (ProxyMinute) {
    var html = '<div class="form-group">\
                <label>代理IP更换频率</label>\
                <select class="form-control" name="ProxyMinute">';
    for (var i in ProxyMinute.menu) {
        var isSelect = ""
        if (ProxyMinute.menu[i] == ProxyMinute.curr[0]) {
            isSelect = " selected";
        }
        ;
        if (ProxyMinute.menu[i] == 0) {
            html += '<option value="' + ProxyMinute.menu[i] + '"' + isSelect + '>' + "不使用代理" + '</option>';
        } else {
            html += '<option value="' + ProxyMinute.menu[i] + '"' + isSelect + '>' + ProxyMinute.menu[i] + ' min</option>';
        }
    }
    ;
    html += '</select></div>';
    return html;
}

var OutTypeHtml = function (OutType) {
    var html = '<div class="form-group"> \
            <label>输出方式</label>\
            <select class="form-control" name="OutType">';
    for (var i in OutType.menu) {
        var isSelect = "";
        if (OutType.curr == OutType.menu[i]) {
            isSelect = " selected";
        }
        ;
        html += '<option value="' + OutType.menu[i] + '"' + isSelect + '>' + OutType.menu[i] + '</option>';
    }
    return html + '</select></div>';
}

var SuccessInheritHtml = function (SuccessInherit) {
    var html = '<div class="form-group"> \
            <label>继承并保存成功记录</label>\
            <select class="form-control" name="SuccessInherit">';

    var True = "";
    var False = "";
    if (SuccessInherit == true) {
        True = " selected";
    } else {
        False = " selected";
    }
    ;

    html += '<option value="true"' + True + '>' + "Yes" + '</option>';
    html += '<option value="false"' + False + '>' + "No" + '</option>';
    return html + '</select></div>';
}

var FailureInheritHtml = function (FailureInherit) {
    var html = '<div class="form-group"> \
            <label>继承并保存失败记录</label>\
            <select class="form-control" name="FailureInherit">';

    var True = "";
    var False = "";
    if (FailureInherit == true) {
        True = " selected";
    } else {
        False = " selected";
    }
    ;

    html += '<option value="true"' + True + '>' + "Yes" + '</option>';
    html += '<option value="false"' + False + '>' + "No" + '</option>';
    return html + '</select></div>';
}

var btnHtml = function (mode, status) {
    if (parseInt(mode) != offline) {
        return '<button type="submit" id="btn-run" class="btn btn-primary" data-type="run">Run</button>';
    }
    switch (status) {
        case _stopped:
            return '<button type="button" id="btn-pause" class="btn btn-warning" onclick="pauseRecover()" disabled="disabled">Pause</button>\
            <button type="submit" id="btn-run" class="btn btn-primary" data-type="run">Run</button>';
        case _stop:
            return '<button type="button" id="btn-pause" class="btn btn-warning" onclick="pauseRecover()" disabled="disabled">Pause</button>\
            <button type="submit" id="btn-run" class="btn btn-danger" data-type="stop" disabled="disabled">Stopping...</button>';
        case _run:
            return '<button type="button" id="btn-pause" class="btn btn-warning" onclick="pauseRecover()" style="display:inline-block;" >Pause</button>\
            <button type="submit" id="btn-run" class="btn btn-danger" data-type="stop">Stop</button>';
        case _pause:
            return '<button type="button" id="btn-pause" class="btn btn-info" onclick="pauseRecover()" style="display:inline-block;" >Go on...</button>\
            <button type="submit" id="btn-run" class="btn btn-danger" data-type="stop">Stop</button>';
    }
}

var logBoxHtml = function (m) {
    if (m == client) {
        return '<div class="box log client">\
              <div class="box-body chat" id="log-box">\
              </div>\
          </div>';
    }

    return '<div id="b" class="split content">\
                <div class="box log">\
                    <div class="box-body chat" id="log-box">\
                    </div>\
                </div>\
            </div>';
};

//生成到</head>的所有代码
var headHTML = function () {
    return '<!DOCTYPE html>\n' +
        '<html lang="en">\n' +
        '<head>\n' +
        '    <meta charset="UTF-8">\n' +
        '    <title>FKAutoSpiderPool（FK自动采集爬虫）</title>\n' +
        '    <!-- Tell the browser to be responsive to screen width -->\n' +
        '    <meta content="width=device-width, initial-scale=1, maximum-scale=1, user-scalable=no" name="viewport">\n' +
        '    <link rel="shortcut icon"\n' +
        '          href="data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAABgAAAAYCAYAAADgdz34AAAAAXNSR0IArs4c6QAAAARnQU1BAACxjwv8YQUAAAAJcEhZcwAADsMAAA7DAcdvqGQAAAD/SURBVEhL7ZUtjgJBEIUHBRIHjnCJVTjEKo4Bd8OgUBjEsslKzrAEQ9gzLO/NdCXVNf0z02P5ki+prn7VJSBQvenLJ/yF/4Vylm9EucHQYB+5JIqESsnO68A3vDSlVxPJ6Tyx5xY68OUkuiaS03lizy2ygQydFww1SihcYpRsYCglCzbw4WSdpGTBHcoc6yR9FozgHsqMyB7vgkgohX6six7BpkEP0y0UdtDeewSbEfihMrusTw2s2eNdkC4LJCNOoDCG9t4j9XMt2P4CCnNo7z1SfziWH8j+uj41rCB71/pUiF5Kz3AGp/Doetre2AdyDoLfmgP8g094gh/QUVUvMGrSh1mUbY0AAAAASUVORK5CYII="\n' +
        '          type="image/x-icon">\n' +
        '    <!-- Bootstrap 3.3.4 -->\n' +
        '    <!--<link href="/public/bootstrap/css/bootstrap.min.css" rel="stylesheet" type="text/css">-->\n' +
        '    <link href="/public/bootstrap/css/bootstrap.min.css" rel="stylesheet" type="text/css">\n' +
        '    <link href="/public/css/default.css" rel="stylesheet" type="text/css">\n' +
        '    <script src="/public/js/jquery.min.js"></script>\n' +
        '    <script src="/public/bootstrap/js/bootstrap.min.js"></script>\n' +
        '    <script src="/public/splitjs/split.js"></script>\n' +
        '</head>'
};

//生成调用splitjs的代码
var splitJSHTML = function () {
    return '<script>\n' +
        '    Split([\'#a\', \'#b\'], {\n' +
        '        direction: \'vertical\',\n' +
        '        gutterSize: 8,\n' +
        '        sizes: [80, 20],\n' +
        '        cursor: \'row-resize\'\n' +
        '    })\n' +
        '\n' +
        '    Split([\'#c\', \'#d\'], {\n' +
        '        sizes: [70, 30],\n' +
        '        gutterSize: 8,\n' +
        '        cursor: \'col-resize\'\n' +
        '    })\n' +
        '</script>'
};
