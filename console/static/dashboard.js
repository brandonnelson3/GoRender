console.log("Dashboard.js Loaded...");



document.addEventListener("DOMContentLoaded", function() {
    'use strict';

    var cameraPositionElement = document.getElementsByClassName('camera_position')[0];
    var cameraForwardElement = document.getElementsByClassName('camera_forward')[0];
    var cameraAngleElement = document.getElementsByClassName('camera_angle')[0];
    var connectionStatusElement = document.getElementsByClassName('connection_status_img')[0];
    var cascade1Element = document.getElementsByClassName('cascade_1')[0];
    var cascade2Element = document.getElementsByClassName('cascade_2')[0];
    var cascade3Element = document.getElementsByClassName('cascade_3')[0];
    var cascade1ShadowElement = document.getElementsByClassName('cascade_shadow_1')[0];
    var cascade2ShadowElement = document.getElementsByClassName('cascade_shadow_2')[0];
    var cascade3ShadowElement = document.getElementsByClassName('cascade_shadow_3')[0];

    var chart = new SmoothieChart({
        millisPerPixel:72,
        grid:{
            fillStyle:'#ffffff',
            strokeStyle:'rgba(119,119,119,0.99)',
            sharpLines:true,
            verticalSections:7
        },
        labels:{
            fillStyle:'#000000'
        },
        minValue:0
    }),
    canvas = document.getElementById('fps_counter'),
    line = new TimeSeries();

    chart.addTimeSeries(line, {lineWidth:2.5,strokeStyle:'#192047',fillStyle:'#303e89'});
    chart.streamTo(canvas, 551);
    chart.addTimeSeries(line);
    
    var ws = null;
    function start(){    
        ws = new WebSocket("ws://localhost:8080/ws");
        ws.onopen = function(){
            console.log('connected!');
            connectionStatusElement.src = "/static/green_light.png";
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

            if (data.type == "timer_fps") {
                line.append(new Date().getTime(), data.value);
            }
        };
        ws.onclose = function(){
            console.log('closed!');

            connectionStatusElement.src = "/static/red_light.png";

            line.append(new Date().getTime(), 0);
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