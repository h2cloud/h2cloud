var h2=require("./h2");
var fs=require("fs");

var args=process.argv;
var argsMap={};
var isRaw=false;
var cache={};
var enableCache=false;
var container="defaultCon";
var patternFile="pattern.txt";

var anotherArgs=[];
for (var i=0; i<args.length; i++) {
    if (args[i].indexOf("-")==0) {
        argsMap[args[i].substr(1)]=true;
    } else {
        anotherArgs.push(args[i]);
    }
}
args=anotherArgs;

if (args.length>2) {
    container=args[2];
}
if (args.length>3) {
    patternFile=args[3];
}
if ("getraw" in argsMap) {
    isRaw=true;
}
if ("cache" in argsMap) {
    enableCache=true;
}

var fList=JSON.parse(fs.readFileSync(patternFile, {encoding: "utf8"}));
var globalTime=0;
(function goList(i, callback) {
    if (i==fList.length) {
        callback();
        return;
    }
    var elem=fList[i];

    var ppath="";
    if (enableCache) {
        var p=elem.lastIndexOf("/");
        if (p>=0) {
            ppath=elem.substring(0, p+1);
            var pfn=elem.substring(p+1);
            if (ppath in cache) {
                elem="/[[SC]"+cache[ppath]+"]/"+pfn;
            }
        }
    }
    h2.getFile(container, elem, function(time, hder) {
        console.log("Fetched", elem, "in", time, "ms");
        globalTime+=time;
        if (enableCache && ppath!="") {
            if ("parent-node" in hder) {
                cache[ppath]=hder["parent-node"];
            }
        }
        process.nextTick(function() {
            goList(i+1, callback);
        });
    }, isRaw);

})(0, function() {
    console.log("average fetch time:", globalTime/fList.length)
    process.exit(0);
});
