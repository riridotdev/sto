# Sto

Sto is a simple CLI tool for creating, managing, and sharing symbolic links to
dotfiles and other configuration files.

## Building
```
git clone https://github.com/riridotdev/sto
cd sto
make
```

## Installation
```
go install https://github.com/riridotdev/sto
```

## Example Usage
##### Initialise a Sto profile
```
sto init
```

##### Manage a file with Sto
```
sto pull [target-path]
```

##### List links for current profile
```
sto list
```

##### Create link for managed item
```
sto push [name]
```

##### List available commands
```
sto --help
```
