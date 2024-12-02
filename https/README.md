# moxy https

## Run moxy on https:

```sh  
cd moxy/https
./gencert
```  

```sh  
sudo cp moxyserver.crt /usr/local/share/ca-certificates/
sudo update-ca-certificates
```

```sh  
moxy -p 9097 -s
```

## Making requests
Request from python:
```sh 
requests.get("https://localhost:9097/api/test", verify=False)
```

Request from java:
```sh 
keytool -importcert -file moxyserver.crt -keystore $JAVA_HOME/lib/security/cacerts -alias my-proxy-cert
```