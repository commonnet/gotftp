#### Trivial File Transfer Protocol Server in Go
##### Setup:
```
$ go get github.com/nalapati/gotftp
```
##### Usage:
```
$ $GOPATH/bin/gotftp <filesystem root> <filesystem tmp> <interface ip4> <port>

All parameters are required:
<filesystem root> The location on the server where files are read from, and where
                  files will be saved to.
<filesystem tmp>  The location where files are staged while they are being written
                  to. This implementation of tftp accepts a write request for a
                  file and stages the data transfer to the <filesystem tmp>
                  location, once the file transfer is complete, moves the file
                  from <filesystem tmp> to <filesystem root>
<interface ip4>   The ip of the interface the tftp server should listen on.
<port>            The port the tftp server should listen on.
```
##### Example:
```
$ $GOPATH/bin/gotftp /tmp/fsroot /tmp/fstmp 127.0.0.1 8000
```
The tftp implementation is per [rfc1350](http://www.ietf.org/rfc/rfc1350.txt). 

##### Testing:
```
$ $GOPATH/bin/gotftp /tmp/fsroot /tmp/fstmp 127.0.0.1 8000

$ echo "hello world" > test.txt
$ sudo apt-get install tftp
$ tftp
tftp> binary
tftp> connect 127.0.0.1 8000
tftp> put test.txt
Sent 12 bytes in 0.0 seconds
tftp> get test.txt
Received 12 bytes in 0.0 seconds
tftp>
```

##### Getting Go Setup (2014-07-19):
```
$ sudo apt-get install golang-go
$ go env

I use bash, I have the following in my .bashrc file.

export GOROOT=/usr/lib/go
export GOPATH=$HOME/go-workspace/gotftp
export PATH=$PATH:${GOPATH//://bin:}/bin
export PATH="$GOROOT:$PATH"

$ source ~/.bashrc
$ go env # validate the GOPATH and GOROOT settings
```

##### Running the unit tests:
```
/usr/bin/go test -v github.com/nalapati/gotftp
```

##### History
1.0 : Basic Implementation responds to wrqs and rrqs, error handling reduces to sending an illegal request for all errors, no retries on failures/timeouts.
