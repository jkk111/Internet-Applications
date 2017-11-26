/*
 * Simple script that calls docker commands to setup N instances of a service
 */

let spawn = require('child_process').spawn;
let express = require('express')
let app = express();

let sessions = {};

let kill = (id) => {
  return new Promise((resolve) => {
    let proc = spawn('docker', [ 'kill', id ]);

    proc.stdout.pipe(process.stdout)
    proc.stderr.pipe(process.stderr)

    proc.on('exit', () => {
      resolve();
    })
  })
}

let create = (image, session) => {
  if(!sessions[session])
    sessions[session] = [];
  return new Promise((resolve) => {
    let proc = spawn('docker', [ 'run', '--net=public_net', '-d', '--expose', '8888', image ]);
    let output = ''
    proc.stdout.on('data', (d) => {
      output += d.toString();
    })

    proc.once('exit', () => {
      let code = output.trim();
      proc = spawn('docker', [ 'inspect', code ]);
      output = '';

      proc.stdout.on('data', (d) => {
        output += d.toString();
      })

      proc.on('exit', () => {
        let d = JSON.parse(output);
        resolve(`http://${d[0].NetworkSettings.IPAddress}:8888`);
      })
    })
  })
}

app.get('/start', async(req, res) => {
  let resp = []
  for(var i = 0; i < req.query.count; i++) {
    resp[i] = await create(req.query.image, req.query.session)
  }
  res.send(resp);
});

app.get('/stop', async(req, res) => {
  let session = sessions[req.query.session];

  console.log(session, sessions, req.query.session)

  if(session) {
    for(var host in session) {
      await kill(host);
    }
  }

  res.send("OK")
})

app.listen(8181)