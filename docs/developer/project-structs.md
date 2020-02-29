# Project Struct

## Introduction

KT-Connect continue the flow parts:

```
                                        Usage Visualization
                                                ^
                                                |
                   +---------------------------------------+
                   |                            |          |
                   |                            |          |
+----------+       |       +----------+   +-----+----+     |
|   ktctl  +---------------+  shadow  |   | Dashboard|     |
+----------+       |       +----------+   +----------+     |
                   |                                       |
                   |                                       |
                   +---------------------------------------+
```

* ktctl: the cli to help user connect to kubernetes cluster
* shadow: the proxy agent that exhange the network request
* dashboard: visualization of the client connect

## Project Directory

```
.
├── cmd
│ ├── ktctl
│ │   └── main.go # The main of ktctl
│ ├── server
│ │   └── main.go # the main of dashboard api
│ └── shadow
│     └── main.go # the main of shadow
├── src # the dashboard front-end
│   ├── components
│   ├── layouts
│   ├── routes
│   ├── store
│   └── utils
└── test
    └── integration
```