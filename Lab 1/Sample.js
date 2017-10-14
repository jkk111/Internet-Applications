let client_1 = [
`HELO test
`,

`AWAIT`,

`JOIN_CHATROOM: primary
CLIENT_IP: 0
PORT: 0
CLIENT_NAME: I AM A BOT
`,
`AWAIT`,

`CHAT: {ROOM_REF}
JOIN_ID: {JOIN_ID}
CLIENT_NAME: I AM A BOT
MESSAGE: Now this is a story all about how
My life got flipped-turned upside down
And I'd like to take a minute
Just sit right there
I'll tell you how I became the prince of a town called Bel-Air

`, `AWAIT`,

`LEAVE_CHATROOM: {ROOM_REF}
JOIN_ID: {JOIN_ID}
CLIENT_NAME: I AM A BOT`,

`AWAIT`,

`KILL_SERVICE
`
]

let client_2 = [
`HELO test
`,

`AWAIT`,

`JOIN_CHATROOM: primary
CLIENT_IP: 0
PORT: 0
CLIENT_NAME: I AM A BOT
`,
`AWAIT`,

`CHAT: {ROOM_REF}
JOIN_ID: {JOIN_ID}
CLIENT_NAME: I AM A BOT
MESSAGE: Now this is a story all about how
My life got flipped-turned upside down
And I'd like to take a minute
Just sit right there
I'll tell you how I became the prince of a town called Bel-Air

`, `AWAIT`
]

module.exports = [ client_1, client_2 ]