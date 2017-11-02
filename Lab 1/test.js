let port = process.argv[2] || 8888;


let spawn = require('child_process').spawn;
let server = spawn('go', [ 'run', 'index.go', port ]);
let listener = spawn('node', [ 'idle_client.js', 'primary', port ]);

listener.stdout.on('data', d => process.stdout.write(d))

// Thanks to race conditions, clients messages are received in random order
let test_client = spawn('node', [ 'client.js' ])

server.on('close', exit_code => {
  if(exit_code)
    throw new Error(`Server exited with code ${exit_code}`)
})