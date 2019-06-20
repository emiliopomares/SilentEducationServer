const WebSocket = require('ws')

const url = "ws://127.0.0.1:9188/";

const conn = new WebSocket(url);

conn.onopen = () => {
  conn.send("shit")
  console.log("WebSocket open")
}

conn.onmessage = (msg) => {
  console.log(msg.data)
}

conn.onerror = error => {
  console.log(`WebSocket error: ${error}`)
}
