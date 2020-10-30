import { EventEmitter } from 'events';

class Socket {
  constructor(
    ws = new WebSocket('ws://localhost:4000'),
    ee = new EventEmitter()
  ) {
    this.ws = ws;
    this.ee = ee;
    ws.onmessage = this.message;
    ws.onopen = this.open;
    ws.onclose = this.close;
  }

  on = (name, func) => {
    this.ee.on(name, func);
  };

  off = (name, func) => {
    this.ee.removeListener(name, func);
  };

  emit = (name, data) => {
    const message = JSON.stringify({ name, data });
    this.ws.send(message);
  };

  message = (e) => {
    // try {
    //   const message = JSON.parse(e.data);
    //   this.ee.emit(message.name, message.data);
    // } catch (err) {
    //   this.ee.emit('error', err);
    // }
    try {
      // console.log(this.ee);
      const message = JSON.parse(e.data);
      this.ee.emit(message.name, message.data);
    } catch (err) {
      this.ee.emit('error', err);
    }
  };

  open = () => {
    this.ee.emit('connect');
  };

  close = () => {
    this.ee.emit('disconnect');
  };
}

export default Socket;
