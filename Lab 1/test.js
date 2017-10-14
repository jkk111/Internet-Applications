let port = process.argv[2] || 8888;


let spawn = require('child_process').spawn;
let server = spawn('go', [ 'run', 'index.go', port ]);
let listener = spawn('node', [ 'idle_client.js', 'primary', port ]);

listener.stdout.on('data', d => process.stdout.write(d))

let test_client = spawn('node', [ 'client.js' ])

test_client.on('close', console.log)