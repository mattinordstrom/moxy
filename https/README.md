# moxy https

To run moxy on https:

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
