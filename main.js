let websocket;

const mediaSource = new MediaSource();
let buffer;
const queue = [];

const video = document.querySelector('video');
video.src = window.URL.createObjectURL(mediaSource);
// video.play();

const mimeCodec = 'video/mp4; codecs="avc1.64001e, mp4a.40.2"';

let connected = false;
let segmentIndex = 0;

const connect = function() {
    if (!connected) {
        websocket = new WebSocket('ws://localhost:9999/ws');
        websocket.binaryType = 'arraybuffer';

        websocket.onopen = function() {
            connected = true;
            console.log("Connected to the web socket");
        };
        websocket.onmessage = function (e) {
            if (typeof e.data !== 'string') {
                if (buffer.updating) {
                    queue.push(e.data);
                } else {
                    buffer.appendBuffer(e.data);
                }
            }
        }
    }
};

mediaSource.addEventListener('sourceopen', function(e) {
    console.log('media source sourceopen');

    connect()

    buffer = mediaSource.addSourceBuffer(mimeCodec);

    buffer.addEventListener('updatestart', function(e) {
        console.log('updatestart: ' + mediaSource.readyState);

        segmentIndex++;
    });
    buffer.addEventListener('updateend', function(e) {
        console.log('updateend: ' + mediaSource.readyState);

        video.play();
    });
    buffer.addEventListener('error', function(e) { console.log('error: ' + mediaSource.readyState); });
    buffer.addEventListener('abort', function(e) { console.log('abort: ' + mediaSource.readyState); });

    buffer.addEventListener('update', function() {
        console.log('buffer update: ' + mediaSource.readyState);
        console.log('buffer update segmentIndex: ', segmentIndex);

        buffer.appendBuffer(queue.shift());
    });

    mediaSource.addEventListener('sourceended', function(e) { console.log('sourceended: ' + mediaSource.readyState); });
    mediaSource.addEventListener('sourceclose', function(e) { console.log('sourceclose: ' + mediaSource.readyState); });
    mediaSource.addEventListener('error', function(e) { console.log('error: ' + mediaSource.readyState); });
}, false);


