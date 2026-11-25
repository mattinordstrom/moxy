# moxy
##### mocking and proxying requests on localhost  
<br>

## Setup:
#### Install go
Follow instructions here: https://go.dev/doc/install  
  

#### Add alias
Add to ~/.zshrc: 
```sh
alias moxy="cd /home/matti/projects/moxy && go run main.go"
```
<br>
  
  
#### Edit config
First time starting moxy a config.yml file will be created
<br>
  
  
## Run:  
#### Run (on default port 9097)  
```sh
$ moxy
```
    
  
#### Run on another port    
```sh
$ moxy -p 9098
```  
  
    
#### Listen on all network interfaces    
By default moxy only accepts connections from this machine (127.0.0.1 and ::1). Use `-a` to make it reachable from other devices, containers or VMs.
```sh
$ moxy -a
```  
  
    
#### See help    
```sh
$ moxy -h
```  
<br>  
    
    
## Admin UI:
Go to http://localhost:9097/moxyadminui
<br>  <br>  
    

    
## Websocket mock:
Any message sent to ws://localhost:9097/moxywsmock will echo broadcast to all connected clients.
Test in terminal:
```sh
$ npm install -g wscat  
$ wscat -c ws://localhost:9097/moxywsmock
```  
or you can use https://github.com/mattinordstrom/moxy_ws_mock   
<br>  
  


## Claude Code skill:
Set up mocks and proxies by asking instead of editing json by hand ("mock GET /api/users/123 as a 500").
See [claude-skill/README.md](claude-skill/README.md) for how to install it.
<br>  
  


## Quick notes:
- Json files will be read for every request (cached for 1.5 sec), so no need to restart moxy after change in json file
- Mocks will always go before proxies.  
- Json files are read from top to bottom and request is processed at first match.
- If no match is found in Json the "Default route" will be used
- URLs will be matched "if contains", so the order and precision in the json urlpart is important
- If payload and payloadFromFile are both defined payload will be used
- .* can be used in urlpart in mocks and proxies
- Want to run on https? Check out the README in the https directory
- Want to copy a request as curl? Click on the small square in the log next to the request (Note that Verbose must be checked)
- Forwarded requests get the target's Host header by default. Set `preserveHost` on a proxy (or `preserveHost: true` under `defaults` in config.yml for the default route) to forward the incoming Host instead, e.g. when the target routes on Host and the client sets it explicitly. A loopback Host (localhost:9097 etc.) is never forwarded
  
<br>  
<br>
______________________________________________
<br>  
<br>

## For developers:
### Run tests:  
```sh
$ go test ./... -v
```  
  
### Run UI tests:  
```sh
$ npm install && npm test
```  

  
### Run lint:  
```sh  
$ golangci-lint run -v
```  
  

  
  
### Precommit hook installation:  
```sh
$ pipx install pre-commit
$ pre-commit install
$ cd ui/tests && npm i
```  
  
  
#### Run precommit manually    
```sh
$ pre-commit run --all-files
```  
