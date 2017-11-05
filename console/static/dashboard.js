console.log("Dashboard.js Loaded...");



document.addEventListener("DOMContentLoaded", function() {
    'use strict';

    var cameraPositionElement = document.getElementsByClassName('camera_position')[0];
    var cameraForwardElement = document.getElementsByClassName('camera_forward')[0];
    var cameraAngleElement = document.getElementsByClassName('camera_angle')[0];

    var ws = null;
    function start(){    
        ws = new WebSocket("ws://localhost:8080/ws");
        ws.onopen = function(){
            console.log('connected!');
        };
        ws.onmessage = function(e){
            var data = JSON.parse(e.data);

            if (data.type == "camera_position") {
                cameraPositionElement.innerHTML = data.value;
            }

            if (data.type == "camera_forward") {
                cameraForwardElement.innerHTML = data.value;
            }

            if (data.type == "camera_angle") {
                cameraAngleElement.innerHTML = data.value;
            }
        };
        ws.onclose = function(){
            console.log('closed!');
            //reconnect now
            check();
        };    
    }    
    function check(){
        if(!ws || ws.readyState == 3) {
            start();
        }
    }
    start();
    setInterval(check, 250);
});