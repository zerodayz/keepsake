var canvas = document.getElementById('canvas');
var ctx = canvas.getContext("2d");
var isDrawing = false;
var strokeColor = '';
var strokes = [];
var items = document.getElementById("items");
var users = document.getElementById("users");

var socket = new WebSocket("wss://" + window.location.hostname + ":3000/ws")
var otherColors = {};
var otherStrokes = {};


function getCookie(cname) {
    var name = cname + "=";
    var decodedCookie = decodeURIComponent(document.cookie);
    var ca = decodedCookie.split(';');
    for(var i = 0; i <ca.length; i++) {
        var c = ca[i];
        while (c.charAt(0) == ' ') {
            c = c.substring(1);
        }
        if (c.indexOf(name) == 0) {
            return c.substring(name.length, c.length);
        }
    }
    return "";
}

board = document.getElementById("board");
canvas.width = board.clientWidth;
canvas.height = "618";


canvas.onmousedown = function (event) {
    isDrawing = true;
    addPoint(event.pageX - this.offsetLeft, event.pageY - this.offsetTop, true);
};
canvas.onmousemove = function (event) {
    if (isDrawing) {
        addPoint(event.pageX - this.offsetLeft, event.pageY - this.offsetTop);
    }
};
canvas.onmouseup = function () {
    isDrawing = false;
};
canvas.onmouseleave = function () {
    isDrawing = false;
};
function addPoint(x, y, newStroke) {
    var p = { x: x, y: y };
    if (newStroke) {
        strokes.push([p]);
    } else {
        strokes[strokes.length - 1].push(p);
    }
    socket.send(JSON.stringify({ kind: MESSAGE_STROKE, points: [p], finish: newStroke }));
    update();
}

function update() {
    ctx.clearRect(0, 0, ctx.canvas.width, ctx.canvas.height);
    ctx.lineJoin = 'round';
    ctx.lineWidth = 4;
    // Draw mine
    ctx.strokeStyle = strokeColor;

    if (items) {
        items.innerHTML = "Active users: "
    }
        drawStrokes(strokes);
    // Draw others'
    var userIds = Object.keys(otherColors);
    for (var i = 0; i < userIds.length; i++) {
        var color = otherColors[userIds[i]];
        var userId = userIds[i];
        if (items) {
            items.innerHTML += "<span style=\"color: " + color + "\">" + userId + " </span>"
        }
        ctx.strokeStyle = otherColors[userId];
        drawStrokes(otherStrokes[userId]);
    }
}

function drawStrokes(strokes) {
    for (var i = 0; i < strokes.length; i++) {
        ctx.beginPath();
        for (var j = 1; j < strokes[i].length; j++) {
            var prev = strokes[i][j - 1];
            var current = strokes[i][j];
            ctx.moveTo(prev.x, prev.y);
            ctx.lineTo(current.x, current.y);
        }
        ctx.closePath();
        ctx.stroke();
    }
}

document.getElementById('clearButton').onclick = function () {
    strokes = [];
    socket.send(JSON.stringify({ kind: MESSAGE_CLEAR }));
    update();
};

socket.onmessage = function (event) {
    var messages = event.data.split('\n');
    for (var i = 0; i < messages.length; i++) {
        var message = JSON.parse(messages[i]);
        onMessage(message);
    }
};

function onMessage(message) {
    switch (message.kind) {
        case MESSAGE_CONNECTED:
            strokeColor = message.color;
            for (var i = 0; i < message.users.length; i++) {
                var user = message.users[i];
                otherColors[user.id] = user.color;
                otherStrokes[user.id] = [];
            }
            update();
            break;
        case MESSAGE_USER_JOINED:
            otherColors[message.user.id] = message.user.color;
            otherStrokes[message.user.id] = [];
            update();
            break;
        case MESSAGE_USER_LEFT:
            delete otherColors[message.userId];
            delete otherStrokes[message.userId];
            update();
            break;
        case MESSAGE_STROKE:
            if (message.finish) {
                otherStrokes[message.userId].push(message.points);
            } else {
                var strokes = otherStrokes[message.userId];
                strokes[strokes.length - 1] = strokes[strokes.length - 1].concat(message.points);
            }
            update();
            break;
        case MESSAGE_CLEAR:
            otherStrokes[message.userId] = [];
            update();
            break;
    }
}
