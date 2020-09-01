# zp
PoE management of the Zyxel GS1900-24HP via command line. 

### Usage
Enable PoE for port 10:
```bash
zp -address 192.168.1.1 -username admin -password s3cr37 -port 10 up
```

Disable PoE for port 10:
```bash
zp -address 192.168.1.1 -username admin -password s3cr37 -port 10 down
```

### Building
```bash 
go build
```