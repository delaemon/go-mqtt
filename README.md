# go-mqtt for osx

## server install
```
brew install Mosquitto
```

## server command
```
brew services start mosquitto
brew services stop mosquitto
```

## config
```
/usr/local/etc/mosquitto/mosquitto.conf
```

## client
```
go run main.go
```
