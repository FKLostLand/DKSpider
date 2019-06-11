package FKDownloaderWebBrowser

/*
* system.args[0] == js
* system.args[1] == url
* system.args[2] == cookie
* system.args[3] == pageEncode
* system.args[4] == userAgent
* system.args[5] == postdata
* system.args[6] == method
* system.args[7] == timeout
 */
const globalPhantomJS string = `
var system = require('system');
var page = require('webpage').create();
var url = system.args[1];
var cookie = system.args[2];
var pageEncode = system.args[3];
var userAgent = system.args[4];
var postdata = system.args[5];
var method = system.args[6];
var timeout = system.args[7];

var ret = new Object();
var exit = function () {
    console.log(JSON.stringify(ret));
    phantom.exit();
};

//输出参数
// console.log("url=" + url);
// console.log("cookie=" + cookie);
// console.log("pageEncode=" + pageEncode);
// console.log("userAgent=" + userAgent);
// console.log("postdata=" + postdata);
// console.log("method=" + method);
// console.log("timeout=" + timeout);

// ret += (url + "\n");
// ret += (cookie + "\n");
// ret += (pageEncode + "\n");
// ret += (userAgent + "\n");
// ret += (postdata + "\n");
// ret += (method + "\n");
// ret += (timeout + "\n");
// exit();

phantom.outputEncoding = pageEncode;
page.settings.userAgent = userAgent;
page.settings.resourceTimeout = timeout;
page.settings.XSSAuditingEnabled = true;

function addCookie() {
    if (cookie != "") {
        var cookies = JSON.parse(cookie);
        for (var i = 0; i < cookies.length; i++) {
            var c = cookies[i];

            phantom.addCookie({
                'name': c.name, /* required property */
                'value': c.value, /* required property */
                'domain': c.domain,
                'path': c.path, /* required property */
            });
        }
    }
}

addCookie();

page.onResourceRequested = function (requestData, request) {

};
page.onResourceReceived = function (response) {
    if (response.stage === "end") {
        // console.log("liguoqinjim received1------------------------------------------------");
        // console.log("url=" + response.url);
        //
        // for (var j in response.headers) {//用javascript的for/in循环遍历对象的属性
        //     // var m = sprintf("AttrId[%d]Value[%d]", j, result.Attrs[j]);
        //     // message += m;
        //     // console.log(response.headers[j]);
        //     console.log(response.headers[j]["name"] + ":" + response.headers[j]["value"]);
        // }
        //
        // console.log("liguoqinjim received2------------------------------------------------");

        //在ret中加入header
        ret["Header"] = response.headers;
    }
};
page.onError = function (msg, trace) {
    ret["Error"] = msg;
    exit();
};
page.onResourceTimeout = function (e) {
    // console.log("phantomjs onResourceTimeout error");
    // console.log(e.errorCode);   // it'll probably be 408
    // console.log(e.errorString); // it'll probably be 'Network timeout on resource'
    // console.log(e.url);         // the url whose request timed out
    // phantom.exit(1);
    ret["Error"] = "onResourceTimeout";
    exit();
};
page.onResourceError = function (e) {
    // console.log("onResourceError");
    // console.log("1:" + e.errorCode + "," + e.errorString);

    if (e.errorCode != 5) { //errorCode=5的情况和onResourceTimeout冲突
        ret["Error"] = "onResourceError";
        exit();
    }
};
page.onLoadFinished = function (status) {
    if (status !== 'success') {
        ret["Error"] = "status=" + status;
        exit();
    } else {
        var cookies = new Array();
        for (var i in page.cookies) {
            var cookie = page.cookies[i];
            var c = cookie["name"] + "=" + cookie["value"];
            for (var obj in cookie) {
                if (obj == 'name' || obj == 'value') {
                    continue;
                }
                if (obj == "httponly" || obj == "secure") {
                    if (cookie[obj] == true) {
                        c += ";" + obj;
                    }
                } else {
                    c += "; " + obj + "=" + cookie[obj];
                }
            }
            cookies[i] = c;
        }
        if (page.content.indexOf("body") != -1) {
            ret["Cookies"] = cookies;
            ret["Body"] = page.content;

            // ret = JSON.stringify(resp);
            exit();
        }
    }
};

page.open(url, method, postdata, function (status) {
});
`
