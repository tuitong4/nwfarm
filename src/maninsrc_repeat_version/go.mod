module main

go 1.19

require nwssh v0.0.0

require (
	golang.org/x/crypto v0.5.0 // indirect
	golang.org/x/sys v0.4.0 // indirect
)

//require "golang.org/x/crypto/ssh"  v0.5.0

replace nwssh v0.0.0 => ../nwssh
