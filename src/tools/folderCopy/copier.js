var h2=require("./h2");
var os=require("os");

var args=process.argv;
if (args.length<5) {
    console.log("node copier [container] [src] [des] [{itfrom}] [{itend}];");
    os.exit(1);
}
var container=args[2];
var srcname=args[3];
var desname=args[4];
var itfrom=null, itend=null;
if (args.length>6) {
    itfrom=-(-args[5]);
    itend=-(-args[6]);
}

if (itfrom==null) {
    console.log("Start moving", srcname, "to", desname);
    h2.moveDir(container, srcname, desname, _END);
} else {
    itCopy(itfrom)();
}

function itCopy(i) {
    if (i>=itend) return _END;
    return function() {
        console.log("Start moving", srcname+i, "to", desname+i);
        h2.moveDir(container, srcname+i, desname+i, itCopy(i+1));
    }
}
function _END() {
    console.log("THE END.");
}
