const websocket = new WebSocket("ws://localhost:9999/ws");
websocket.binaryType = "arraybuffer";

websocket.addEventListener("open", function (e) {
    console.log('socket open called');
});

// var FILE = 'test.webm';
var NUM_CHUNKS = 5;
var video = document.querySelector('video');

if (!window.MediaSource) {
    alert('The MediaSource API is not available on this platform');
}

var mediaSource = new MediaSource();

// document.querySelector('[data-num-chunks]').textContent = NUM_CHUNKS;

video.src = window.URL.createObjectURL(mediaSource);

let messageCounter = 0
const maxMessages = 10

mediaSource.addEventListener('sourceopen', function() {
    // var sourceBuffer = mediaSource.addSourceBuffer('video/webm; codecs="vorbis,vp8"');

    // store the buffers until you're ready for them
    var queue = [];
    var sourceBuffer = mediaSource.addSourceBuffer('video/mp4; codecs="avc1.64001e, mp4a.40.2"');
    // console.log(sourceBuffer);

    console.log('MediaSource readyState: ' + this.readyState);

    websocket.addEventListener("message", function (e) {
        console.log('messageCounter', messageCounter)

        // if (messageCounter % 10 === 0) {
        //      console.log('*********************************** messageCounter', messageCounter)
        // }

        // if (messageCounter < maxMessages) {
            // get('ws://localhost:9999/ws', function(uInt8Array) {
            // const myBlob = new Uint8Array(e.data)
            // var file = new Blob([e.data], {type: 'video/mp4'});
            // var chunkSize = Math.ceil(file.size / NUM_CHUNKS);

            // console.log('Number of chunks: ' + NUM_CHUNKS);
            // console.log('Chunk size: ' + chunkSize + ', total size: ' + file.size);

            // if (!sourceBuffer.updating) {
                // console.log('**** sourceBuffer.buffered', sourceBuffer.buffered);

                sourceBuffer.appendBuffer(e.data);
            // }

            // Slice the video into NUM_CHUNKS and append each to the media element.
            // var i = 0;

            // (function readChunk_(i) { // eslint-disable-line no-shadow
            //     var reader = new FileReader();
            //
            //     // Reads aren't guaranteed to finish in the same order they're started in,
            //     // so we need to read + append the next chunk after the previous reader
            //     // is done (onload is fired).
            //     reader.onload = function(e) {
            //         console.log('**** mediaSource.readyState READY STATE', mediaSource.readyState)
            //
            //         console.log('**** e.target.result: ', e.target.result);
            //         console.log('**** new Uint8Array(e.target.result): ', new Uint8Array(e.target.result));
            //
            //         console.log('mediaSource.readyState READY STATE', mediaSource.readyState)
            //         sourceBuffer.appendBuffer(new Uint8Array(e.target.result));
            //         console.log('Appending chunk: ' + i);
            //
            //         if (i === NUM_CHUNKS - 1) {
            //             // sourceBuffer.addEventListener('update', function() {
            //             //     // if (!sourceBuffer.updating && mediaSource.readyState === 'open') {
            //             //     //     mediaSource.endOfStream();
            //             //     // }
            //             // });
            //         } else {
            //             if (video.paused) {
            //                 // video.play(); // Start playing after 1st chunk is appended.
            //             }
            //             readChunk_(++i);
            //         }
            //     };
            //
            //     var startByte = chunkSize * i;
            //     var chunk = file.slice(startByte, startByte + chunkSize);
            //
            //     reader.readAsArrayBuffer(chunk);
            // })(i); // Start the recursive call by self calling.
            // messageCounter++
        // } else if (messageCounter === maxMessages) {
        //     // ++messageCounter
        //
        //     console.log('START PLAYING....')
        //
        //     // mediaSource.endOfStream()
        //     // video.play()
        // }

        messageCounter++
    });
}, false);

mediaSource.addEventListener('sourceended', function() {
    console.log('MediaSource readyState: ' + this.readyState);
}, false);

// function get(url, callback) {
//     var xhr = new XMLHttpRequest();
//     xhr.open('GET', url, true);
//     xhr.responseType = 'arraybuffer';
//     xhr.send();
//
//     xhr.onload = function() {
//         if (xhr.status !== 200) {
//             alert('Unexpected status code ' + xhr.status + ' for ' + url);
//             return false;
//         }
//         callback(new Uint8Array(xhr.response));
//     };
// }

// function log(message) {
//     document.getElementById('data').innerHTML += message + '<br /><br />';
// }