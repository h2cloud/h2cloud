var http=require("http");
var fs=require("fs");

function moveDir(container, filepath, desfilepath, callback) {
    var options = {
        hostname: 'controller',
        port: 9144,
        path: escape('/fs/'+container+filepath),
        method: 'PATCH',
        headers: {
            "C-Destination": desfilepath,
        },
    };

    var req = http.request(options, function(res) {
        if (res.statusCode>=300) {
            console.log("None 2xx code:", res.statusCode);
            callback(1);
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

exports.moveDir=moveDir;
