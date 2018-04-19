# golang networking
Only 1 on 1 connection works atm

Interface:

Init([]Remote)
"data from remote device" <- Remote.Receive
"true/false when device connects/disconnects" <- Remote.Connected
Remote.Send("data to remote device")    /* only ints work atm */

Get_remotes() returns int number of remotes initialized
