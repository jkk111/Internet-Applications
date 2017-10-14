let net = require('net');
let colors = require('colors')

if(process.argv.length < 4) {
  console.error("Usage node idle_client.js {room} {port}".red)
  process.exit(1);
}

let room = process.argv[2];
let port = process.argv[3];

console.info("Attempting to join room: %s on port %d".yellow, room, port);

let hello = () => {
  return 'HELO listener\n'
}

let join = (room) => {
  return `JOIN_CHATROOM: ${room}\n` +
  'CLIENT_IP: 0\n' +
  'PORT: 0\n' +
  'CLIENT_NAME: idle_listener\n'
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

let log_message = (d) => {
  d = d.toString();

  let message = parse_message(d);

  if(message["MESSAGE:"]) {
    let date = new Date() + "";
    date = date.yellow;
    console.log(date)
    let m = (message['CLIENT_NAME:'] + ":").green
    m += " " + message["MESSAGE:"]
    console.log(m)
    console.log()
  }
}

let socket = net.connect(port, 'localhost', () => {
  console.info("Connected to port: %d\n".yellow, port)
  socket.write(hello())
  socket.write(join(room))

  socket.on('data', log_message);
});