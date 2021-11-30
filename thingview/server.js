// Launch dashboard using node 
// $ node ./server.js &
const express = require('express');
const bodyParser = require('body-parser')
const cors = require('cors')
const path = require('path');
const fs = require('fs')
const http = require('http')
const https = require('https')

const privateKey = fs.readFileSync('./certs/serverKey.pem')
const certificate = fs.readFileSync('./certs/serverCert.pem')

var credentials = {key: privateKey, cert:certificate}


const app = express();
//app.use(cors)
app.use(express.static(__dirname+'/dist'));


var corsOptions = {
  // origin: 'https://localhost',
  // optionsSuccessStatus: 200 // some legacy browsers (IE11, various SmartTVs) choke on 204
}

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
console.log("Listening on port " + (process.env.PORT || 8443) + " and 8080" ) 


var httpServer = http.createServer(app)
var httpsServer = https.createServer(credentials,app)
httpServer.listen(8080)
httpsServer.listen(8443)
//app.listen(process.env.PORT || 8080);


