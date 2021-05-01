# plot-util

A simple tool for chia plotting.

## Usage

This tool can be used to query remote farmer plot status, and fetch plot files from remote server.

Add local ssh public key to all servers' $HOME/.ssh/authorized_keys first:

```
ssh-copy-id $user@$server
```

Compile, configure, and run:

```bash
mkdir bin
go build -o bin/plot main.go

cp conf/example.hosts.yaml conf/hosts.yaml
# modify your config
vim conf/hosts.yaml

# only query status
bin/plot --debug

# only fetch latest plot file
bin/plot --debug --fetch

# fetch all the plot files
bin/plot --debug --fetch --loop
```

## TODO

- [ ] add sshpass for ssh password authenticating.

- [ ] add pull plot files concurrently.

- [ ] add query plot progress.

- [ ] add telegram bot to report plot progress and status.
