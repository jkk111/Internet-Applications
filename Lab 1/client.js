let net = require('net')
let clients = require(__dirname + '/Sample.js')

class Socket {
  constructor(port, host) {
    this._sock = net.connect(8888, host, () => {
      this.data = {}
      this.connected = true;
    })

    this._sock.on('error', (e) => {
      if(e.code !== 'ECONNRESET') {
        console.log('Unexpected Error!')
      }
    })
  }

  parse_message(d) {
    let fields = {};
    d = d.toString();

    let get_type = (d) => {
      return d.slice(0, d.indexOf(' '))
    }

    let get_data = (d, term) => {
      return d.slice(0, d.indexOf(term))
    }

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
    this.data = Object.assign(this.data, fields);
    return fields
  }

  format_message(d) {
    let match = d.match(new RegExp('{(.*?)}'));
    while(match) {
      let key = match[1];
      let value = this.data[key];

      if(key && value) {
        d = d.replace(`{${key}}`, value);
      } else {
        break;
      }

      match = d.match(new RegExp('{(.*?)}'))
    }
    return d;
  }

  ready() {
    return new Promise((resolve) => {
      if(this.connected) {
        resolve();
      } else {
        this._sock.once('connect', () => {
          resolve();
        })
      }
    })
  }

  async send(message) {
    await this.ready();
    let formatted = this.format_message(message);
    this._sock.write(formatted);
  }

  async receive() {
    await this.ready();
    return new Promise((resolve) => {
      this._sock.once('data', (d) => {
        this.parse_message(d);
        resolve(d);
      })
    })
  }
}

let Receive = socket => {
  return new Promise((resolve) => {
    socket.once('data', (d) => {
      let msg = d.toString()
      console.log("MESG", msg.toString())
      resolve(msg)
    })
  })
}

let run = async() => {
  for(var messages of clients) {
    (async (messages) => {
      let socket = new Socket(8888, 'localhost');
      for(var message of messages) {
        if(message === 'AWAIT') {
          let message = await socket.receive();
          console.log(message.toString())
        } else {
          await socket.send(message);
        }
      }
    })(messages)
  }
}

run()