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
<br>
  
  
## Admin UI:
Go to http://localhost:9097/moxyadminui
  

  
## Quick notes:
- Json files will be read for every request (cached 2 sec), so no need to restart moxy after change in json file
- Mocks will always go before proxies.  
- Json files are read from top to bottom and request is processed at first match.
- If no match is found in Json the "Default route" will be used
- URLs will be matched "if contains", so the order and precision in the json urlpart is important
- If payload and payloadFromFile are both defined payload will be used
- .* can be used in urlpart in mocks and proxies
  
<br>  
<br>
_______________________
<br>  
<br>

## For developers:
### Run tests:  
```sh
$ go test ./... -v | sed -e 's/PASS/\x1b[32m&\x1b[0m/' -e 's/FAIL/\x1b[31m&\x1b[0m/'
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
