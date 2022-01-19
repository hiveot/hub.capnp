// Launch dashboard using node from the hub home folder (~/bin/hub
//  >  node ./bin/server.js &
// this will serve thingview on port 8443
const express = require('express');
const path = require('path');
const fs = require('fs')
const https = require('https')

const privateKey = fs.readFileSync('./certs/serverKey.pem')
const certificate = fs.readFileSync('./certs/serverCert.pem')

var credentials = {key: privateKey, cert:certificate}

const app = express();


const publicPath = path.join(__dirname, 'thingview');
app.use(express.static(publicPath));

// for browserHistory:
// https://github.com/reactjs/react-router/blob/1.0.x/docs/guides/basics/Histories.md
app.get('/favicon*', function(req, resp) {
  resp.sendFile(path.resolve(publicPath, 'favicon.png'));
});
app.get('/', function(req, resp) {
  resp.sendFile(path.resolve(publicPath, 'index.html'));
});
// For Vue (and react) all requests lead to index.html
// app.get('*', function(req, resp) {
//   console.log(".get", req.path, req.body, req.headers )
//   // resp.sendFile(path.resolve(publicPath, 'index.html'));
//   let filePath = path.resolve(path.join(publicPath, req.path));
  
//   resp.sendFile(filePath);
// });

console.log("Service %s on port %s", publicPath, (process.env.PORT || 8443) )
let httpsServer = https.createServer(credentials,app)
httpsServer.listen(8443)


