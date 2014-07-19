Trivial File Transfer Protocol Server Implementation in Go
----------------------------------------------------------

$ go get github.com/nalapati/gotftp
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

Example:
$$GOPATH/bin/gotftp /tmp/fsroot /tmp/fstmp 127.0.0.1 8000



