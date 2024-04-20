import ws from 'k6/ws';
import { check } from 'k6';
import { Counter } from 'k6/metrics';

// A counter for the number of messages sent
const counterSent = new Counter('ws_messages_sent');
const counterRecived = new Counter('ws_messages_recived');

const expectedResponse = '{"result": "success", "message": "The server is up and running."}';

export default function () {
  const url = 'ws://localhost:8080/';
  const params = { tags: { my_tag: 'hello' } };

  let counter = 100;

  const res = ws.connect(url, params, function (socket) {
    socket.on('open', () => {
      console.log('connected');
      // Send as many messages as possible

      const iterations = counter;
      for (let i = 0; i < iterations; i++) {
        socket.send('Hello, server!');

        counterSent.add(1);
      }
    });

    socket.on('message', (data) => {
      check(data, { 'Got Expected message': (d) => d === expectedResponse });
      counterRecived.add(1);
      counter--; 
      if (counter == 0) {
        socket.close();
      }
    });
    socket.on('close', () => console.log('disconnected'));
  });

  check(res, { 'status is 101': (r) => r && r.status === 101 });
}
