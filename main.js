let websocket;

const mediaSource = new MediaSource();
let buffer;
const queue = [];

const video = document.querySelector('video');
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
                    console.log("ELSE");
                    buffer.appendBuffer(e.data);
                }
            }
        }
    }
};

mediaSource.addEventListener('sourceopen', function(e) {
    console.log('media source sourceopen');

    connect()
    // video.play();

    buffer = mediaSource.addSourceBuffer(mimeCodec);

    buffer.addEventListener('updatestart', function(e) { console.log('updatestart: ' + mediaSource.readyState); });
    buffer.addEventListener('updateend', function(e) { console.log('updateend: ' + mediaSource.readyState); });
    buffer.addEventListener('error', function(e) { console.log('error: ' + mediaSource.readyState); });
    buffer.addEventListener('abort', function(e) { console.log('abort: ' + mediaSource.readyState); });

    buffer.addEventListener('update', function() {
        console.log('buffer update: ' + mediaSource.readyState);

        if (queue.length > 0 && !buffer.updating) {
            buffer.appendBuffer(queue.shift());
        }
    });

    mediaSource.addEventListener('sourceended', function(e) { console.log('sourceended: ' + mediaSource.readyState); });
    mediaSource.addEventListener('sourceclose', function(e) { console.log('sourceclose: ' + mediaSource.readyState); });
    mediaSource.addEventListener('error', function(e) { console.log('error: ' + mediaSource.readyState); });
}, false);


