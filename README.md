Telemetry Platform (PXE Diagnostics)

This project is a simple telemetry system written in Go.
It is designed to run diagnostic agents on PXE-booted machines and collect heartbeat information on a central server.

The goal of this project is to evolve into a lightweight infrastructure diagnostics and monitoring platform.


The project currently consists of two components:

Server

A minimal HTTP API server that listens for heartbeat messages.

Location:

server/main.go

Responsibilities:

Listen on TCP port (default :8080)

Expose endpoint:

POST /heartbeat

Receive JSON heartbeat messages

Log received machine status

Agent (Client)

A lightweight daemon that runs inside a PXE-booted Alpine Linux image.

Location:

agent/main.go

Responsibilities:

Read configuration from:

/etc/heartbeat-agent/config.json

Collect basic system identity:

Hostname

IPv4 addresses

Send heartbeat periodically to the server




I AM RESISTING THE URGE TO NAME EVERY FILE AND FUNCTION MEME NAMES.