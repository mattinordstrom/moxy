# moxy
##### mocking and proxying requests on localhost  
  

## Setup:
#### Install go
Follow instructions here: https://go.dev/doc/install  
  
##  
#### Add alias
Add to ~/.zshrc: 
```sh
alias moxy="cd /home/matti/projects/moxy && go run main.go"
```

##  
#### Dont track config
```sh
git update-index --assume-unchanged config.yml
```  
  
##  
##  
## Run:  
#### Run (on default port 9097)  
```sh
$ moxy
```
    
##  
#### Run on another port    
```sh
$ moxy -p 9098
```
  
##  
##  
## Edit proxies and mocks:
Edit mocks: [mockdef.json]  
Edit proxies: [proxydef.json]  

##  
##  
## Quick notes:
- Json files will be read for every request, so no need to restart moxy after change in json file
- Mocks will always go before proxies.  
- Json files are read from top to bottom and request is processed at first match.
- If no match is found in Json the "Default route" will be used
- / will be trimmed from all mock URLs (/api/test/ will match api/test)
  

##
##
## Admin UI:
Go to http://localhost:9097/moxyadminui
___

  
##  
##  
## Run tests:  
```sh
$ go test ./... -v
```  
  

##  
##  
## Run lint:  
```sh  
$ golangci-lint run -v
```  
  

##  
##  
## Precommit hook:  
```sh
$ pip3 install pre-commit
$ pre-commit install
```  
  
##  
#### Run precommit manually    
```sh
$ pre-commit run --all-files
```  
  
[mockdef.json]: <https://github.com/mattinordstrom/moxy/blob/main/mockdef.json>
[proxydef.json]: <https://github.com/mattinordstrom/moxy/blob/main/proxydef.json>
