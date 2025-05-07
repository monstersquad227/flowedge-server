# flowedge-server

## interface

```text
GET http://127.0.0.1:8080/execute
```

## params

### containerList 

|name|require|
|-|-|
|agent_id|true|
|command|true|

### containerRemove

|name|require|
|-|-|
|agent_id|true|
|command|true|
|container_id|true|

### containerStop

|name|require|
|-|-|
|agent_id|true|
|command|true|
|container_id|true|

### containerCreate

|name|require|
|-|-|
|agent_id|true|
|command|true|
|image|true|

### containerStart

|name|require|
|-|-|
|agent_id|true|
|command|true|

### containerDragon

|name|require|
|-|-|
|agent_id|true|
|command|true|
|image|true|

### imagePull

|name|require|
|-|-|
|agent_id|true|
|command|true|
|image|true|

### imageList

|name|require|
|-|-|
|agent_id|true|
|command|true|