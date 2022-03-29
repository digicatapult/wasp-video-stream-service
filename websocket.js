//ffmpeg -f x11grab -s 1920x1080 -r 25 -i :0.0+0,0 -f webm -codec:v libvpx -quality good -cpu-used 0 -b:v 600k -maxrate 600k -bufsize 1200k -qmin 10 -qmax 42 -vf scale=-1:480 -threads 4 http://localhost:8081/test

var websocket; // = new WebSocket('ws://localhost:9999/ws');

var mediaSource = new MediaSource();
var buffer;
var queue = [];

var video = document.querySelector('video');
video.src = window.URL.createObjectURL(mediaSource);

const mimeCodec = 'video/mp4; codecs="avc1.64001e, mp4a.40.2"';

let connected = false;

const connect = function() {
    // let socketStartTime = performance.now();
    if (!connected) {
        websocket = new WebSocket('ws://localhost:9999/ws');
        websocket.binaryType = 'arraybuffer';

        websocket.onopen = function() {
            connected = true;
            console.log("Connected to the web socket");
            // capture();
        };
        websocket.onmessage = function (e) {
            // let blob = m.data;
            // renderImage(dCtx, blob);
            // const now = performance.now();
            // socketStartTime = now;

            console.log('websocket message e.data: ' + e.data);

            if (typeof e.data !== 'string') {
                console.log("typeof e.data !== 'string'", e.data);

                if (buffer.updating || queue.length > 0) {
                    console.log("buffer.updating || queue.length > 0");

                    queue.push(e.data);
                } else {
                    console.log("**** ELSE");
                    buffer.appendBuffer(e.data);
                }
            }
        }
    }
};

mediaSource.addEventListener('sourceopen',  async function(e) {
    console.log('media source sourceopen');

    connect()
    // await video.play();

    buffer = mediaSource.addSourceBuffer(mimeCodec);

    buffer.addEventListener('updatestart', function(e) { console.log('updatestart: ' + mediaSource.readyState); });
    // buffer.addEventListener('update', function(e) { console.log('update: ' + mediaSource.readyState); });
    buffer.addEventListener('updateend', async function(e) {
        console.log('updateend: ' + mediaSource.readyState);

        // if (!buffer.updating) {
        //     buffer.appendBuffer(queue.shift());
            // buffer.endOfStream();
            // await video.play()
        // }

        // queue.shift()
    });
    buffer.addEventListener('error', function(e) { console.log('error: ' + mediaSource.readyState); });
    buffer.addEventListener('abort', function(e) { console.log('abort: ' + mediaSource.readyState); });

    buffer.addEventListener('update', function() { // Note: Have tried 'updateend'
        console.log('***** buffer update: ' + mediaSource.readyState);

        // console.log('buffer update e: ' + e.);

        // queue.push(e.)
        if (!buffer.updating) {
            console.log('buffer update !buffer.updating');

            buffer.appendBuffer(queue.shift());
        //     buffer.endOfStream();
        //     await video.play();
        } else {
            console.log('buffer update buffer.updating');
        }
    });

    mediaSource.addEventListener('sourceopen', function(e) { console.log('sourceopen: ' + mediaSource.readyState); });
    mediaSource.addEventListener('sourceended', function(e) { console.log('sourceended: ' + mediaSource.readyState); });
    mediaSource.addEventListener('sourceclose', function(e) { console.log('sourceclose: ' + mediaSource.readyState); });
    mediaSource.addEventListener('error', function(e) { console.log('error: ' + mediaSource.readyState); });
}, false);


