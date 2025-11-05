# Overview
**SCPI** is a standard syntax for remotely controlling test and measurement devices.
This repository contains 2 SCPI clients: a CLI and a web UI.

**Sclipi** is a command line tool for sending **SCPI** commands.
**Scpir** is a web ui for sending **SCPI** commands.

Features:
-   Auto-completion support!
-   History support (Up key cycles through previous commands)
-   Interactive and non-interactive operation
-   Saving command history to file
-   Running scripts from file
-   Copying response to clipboard
-   Copying entire session to clipboard

# Sclipi

## Interactive-Mode

By default **Sclipi** runs in interactive mode. This mode drops you into a shell where auto-completion and custom
commands are enabled.

You will be presented with a prompt asking for the IP address of the target instrument. Auto-completion will assist by
suggesting the first 3 octets of each of the network interfaces on your machine.

Once the connection is established, an auto-completion prompt will appear with all of the available options. Type any of
these to get started:

-   `:` and `*`: SCPI commands
-   `-`: Actions (Show history, Save to script, Run script, Copy result to clipboard, etc)
-   `$`: Shell Passthrough (e.g. `$clear` to clear the terminal)
-   `?`: Help
-   `quit` or `exit`: Exit the shell

The behavior can be tweaked with various arguments, including:

-   `-a|--address <ip-address|hostname>`: Connect to instrument this address (skips IP address prompt)
-   `-p|--port <port>`: Change target SCPI socket port from the default 5025
-   `-q|--quiet`: Suppress most output to reduce clutter
-   Various `--*-color` options: Change the default color of various elements inside the shell

## Non-Interactive Mode

If you know what commands you want to run (or have them saved to a file) and don\'t want to drop into the interactive
shell, there are two command line arguments to run SCPI command directly:

-   `-c|--command <scpi-command>`: Run the provided command and print the result, if any
-   `-f|--file <file-path>`: Run each of the commands in a newline-delimited text file sequentially, printing the
    results, if any

Both non-interactive arguments require that the address of the instrument is also provided using `-a`.

# For Sclipi Developers

## Simulated Instruments

If **Sclipi** is run from a directory containing a SCPI.txt file, it supports parsing that file into a list of supported
SCPI commands. The file must be in the `:SYSTem:HELP:HEADers?` format.

Simulation mode can be triggered two ways: -The `-s|--simulate` argument -Typing `simulated` into the `-a|--address`
argument or the IP Address interactive prompt

# Scpir

## Running Scpir using Docker

Docker images are available at ghcr.io/bhutch29/scpir:master. See the provided `docker-compose.yml` file for a reference on how to set environment variables, etc.

## Running Scpir for development

Start the backend:
```bash
just run-server
```

Start the frontend:
```bash
just serve-web
```

The web interface will be available at `http://localhost:4200`

## Query Parameters

You can pre-configure the connection by passing query parameters in the URL:

-   `?address=<ip-address|hostname>`: Set the instrument address
-   `?port=<port>`: Set the SCPI port (default: 5025)

Example: `http://localhost:8080?address=192.168.1.100&port=5025`
