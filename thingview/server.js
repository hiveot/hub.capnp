// Launch dashboard using node 
// $ node ./server.js &
const express = require('express');
const bodyParser = require('body-parser')
const path = require('path');
const app = express();

app.use(express.static(__dirname+'/dist'));

app.get('/ping', function (req, res) {
 return res.send('pong');
});

// for browserHistory:
// https://github.com/reactjs/react-router/blob/1.0.x/docs/guides/basics/Histories.md
app.get('/favicon*', function(req, res) {
  res.sendFile(path.resolve(__dirname, 'dist', 'favicon.png'));
});
app.get('*', function(req, res) {
  res.sendFile(path.resolve(__dirname, 'dist', 'index.html'));
});

// port must match the proxy line in the frontend package.json
console.log("Listening on port " + (process.env.PORT || 8080))
app.listen(process.env.PORT || 8080);