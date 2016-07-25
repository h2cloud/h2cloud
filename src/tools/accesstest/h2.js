var http=require("http");
var fs=require("fs");

function getFile(container, path, callback, isRaw) {
    var options = {
        hostname: 'controller',
        port: 9144,
        path: escape((isRaw?'/exp/rawget/':'/io/')+container+path),
        method: 'GET',
    };

    var startTime=Date.now();
    var req = http.request(options, function(res) {
        var timeForHeader=Date.now()-startTime;
        if (res.statusCode>=300) {
            console.log("None 2xx code:", res.statusCode);
            return;
        }
        res.setEncoding('utf8');
        res.on('data', function (chunk) {
            // empty
        });
        res.on('end', function() {
            callback(timeForHeader, res.headers);
        });
    });
    req.on('error', function(e) {
        console.log('problem with request: ' + e.message);
        return;
    });
    req.end();
}

function makeDir(container, dir, callback) {
    var options = {
        hostname: 'controller',
        port: 9144,
        path: escape('/fs/'+container+dir),
        method: 'PUT',
    };

    var req = http.request(options, function(res) {
        if (res.statusCode>=300) {
            console.log("None 2xx code:", res.statusCode);
            return;
        }
        res.setEncoding('utf8');
        res.on('data', function (chunk) { });
        res.on('end', function() {
            callback();
        });
    });
    req.on('error', function(e) {
        console.log('problem with request: ' + e.message);
        return;
    });
    req.end();
}

function uploadFile(container, path, realPath, callback) {
    var options = {
        hostname: 'controller',
        port: 9144,
        path: escape('/io/'+container+path),
        method: 'PUT',
    };

    var req = http.request(options, function(res) {
        if (res.statusCode>=300) {
            console.log("None 2xx code:", res.statusCode);
            return;
        }
        res.setEncoding('utf8');
        res.on('data', function (chunk) { });
        res.on('end', function() {
            callback();
        });
    });

    req.on('error', function(e) {
        console.log('problem with request: ' + e.message);
        return;
    });
    var buf=fs.readFileSync(realPath);
    req.write(buf);
    req.end();
}

var escapeChar="<>\"`\r\nt{}|\\^' ";
var descape=[];
for (var i=0; i<escapeChar.length; i++) {
    descape.push(encodeURIComponent(escapeChar[i]));
}
function escape(str) {
    var ret="";
    for (var i=0; i<str.length; i++) {
        var pos=escapeChar.indexOf(str[i]);
        if (pos>=0) {
            ret+=descape[pos];
        } else {
            ret+=str[i];
        }
    }

    return ret;
}

exports.makeDir=makeDir;
exports.uploadFile=uploadFile;
exports.getFile=getFile;
