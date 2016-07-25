var fs=require("fs");
var path=require("path");
var h2=require("./h2");

var args=process.argv;
var rootPath=".";
var container="defaultCon";
if (args.length>2) {
    rootPath=args[2];
}
if (args.length>3) {
    container=args[3];
}

function traverse(directory, onDir, onFile, virtualBaseDirectory) {
    if (virtualBaseDirectory==undefined) {
        virtualBaseDirectory="/"
    }
    var fileList=fs.readdirSync(directory);
    for (var i=0; i<fileList.length; i++) {
        var realPath=path.join(directory, fileList[i]);
        var virtualPath=path.join(virtualBaseDirectory, fileList[i]);
        var stat=fs.statSync(realPath);
        if (stat.isFile()) {
            onFile(realPath, virtualPath.replace(/\\/g, "/"));
        } else if (stat.isDirectory()) {
            onDir(realPath, virtualPath.replace(/\\/g, "/"));
            traverse(realPath, onDir, onFile, virtualPath);
        }
    }
}

var actionLog=[];
traverse(rootPath, function(r, v) {
    actionLog.push(["dir", r, v]);
}, function(r, v) {
    actionLog.push(["file", r, v]);
});

function demo(r, v, callback) {
    callback();
}

(function goList(i, callback) {
    if (i==actionLog.length) {
        callback();
        return;
    }
    var elem=actionLog[i];
    if (elem[0]==="dir") {
        console.log("Making dir:", elem[2]);
        h2.makeDir(container, elem[2], function() {
            process.nextTick(function() {
                goList(i+1, callback);
            });
        });
    } else if (elem[0]==="file") {
        console.log("Uploading file:", elem[2]);
        h2.uploadFile(container, elem[2], elem[1], function() {
            process.nextTick(function() {
                goList(i+1, callback);
            });
        });
    }
})(0, function() {
    process.exit(0);
});
