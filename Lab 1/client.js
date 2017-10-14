let net = require('net')
let clients = require(__dirname + '/Sample.js')

let data = {};

let set_vars = (d) => {
  let match = d.match(new RegExp('{(.*?)}'));
  while(match) {
    let key = match[1];
    let value = data[key];

    if(key && value) {
      d = d.replace(`{${key}}`, value);
    } else {
      break;
    }

    match = d.match(new RegExp('{(.*?)}'))
  }
  return d;
}

let get_type = (d) => {
  return d.slice(0, d.indexOf(' '))
}

let get_data = (d, term) => {
  return d.slice(0, d.indexOf(term))
}

let parse_message = (d) => {
  let fields = {};
  while(d.length) {
    let type = get_type(d);
    let term = "\n";

    d = d.slice(type.length + 1);

    if(type === "MESSAGE:") {
      term += "\n"
    }

    let data = get_data(d, term)
    d = d.slice(data.length + term.length)

    fields[type] = data;
  }
  return fields
}

let Receive = socket => {
  return new Promise((resolve) => {
    socket.once('data', (d) => {
      let msg = d.toString()
      console.log("MESG", msg)
      resolve(msg)
    })
  })
}

let run = async() => {
  for(var messages of clients) {
    ((messages) => {
      let socket = net.connect(8888, 'localhost', async() => {
        for(var message of messages) {
          if(message === 'AWAIT') {
            let resp = await Receive(socket);
            let message = parse_message(resp)
            console.log(message)
            data = Object.assign(data, message)
          } else {
            socket.write(set_vars(message))
          }
        }
      })
    })(messages)
  }
}

run()