let express = require('express')
let multer = require('multer')
let app = express();
var storage = multer.memoryStorage()
let upload = multer({ storage: storage });
let escomplex = require('typhonjs-escomplex');

let request = require('request-promise')

let assigned_nodes = [];
let pending_jobs = [];

let next_node = 0;

app.post('assign', upload.single(), (req, res) => {
  assigned_nodes = req.body.nodes;
  res.send("OK")
});

let dumb_complexity = (s) => {
  return (s.split(/\s/g)).length
}

app.post('/analyze', upload.array('files'), async(req, res) => {
  let compl = {};
  let pending_results = []
  let send_to_worker = (f) => {
    let node = assigned_nodes[(next_node++) % assigned_nodes.length]
    return request({
      url: `http://${node}/analyze`,
      formData: {
        files: [
          {
            value: f.buffer,
            options: {
              filename: f.filename,
              contentType: f.mimetype
            }
          }
        ]
      }
    })
  }

  if(req.files.length > 1 && assigned_nodes.length) {
    pending_jobs = [...pending_jobs, ...req.files];
  } else {
    let file = req.files[0];
    if(file.mimetype === 'application/javascript') {
      let d = file.buffer.toString();
      compl[file.originalname] = {
        type: 'cyclomatic',
        value: escomplex.analyzeModule(d).methodAggregate.cyclomatic
      }
    } else {
      compl[file.originalname] = {
        type: 'simple',
        value: dumb_complexity(d)
      }
    }
  }

  while(pending_jobs.length) {
    let job = pending_jobs.pop();
    pending_results.push(send_to_worker(job));
  }

  for(var i = 0; i < pending_results.length; i++) {
    let result = await pending_results[i];
    compl[result.filename] = result.result
  }

  res.send({
    complexity: compl
  })
});

app.listen(process.env.port || 8080);