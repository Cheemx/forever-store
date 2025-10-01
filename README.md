# Forever-Store

## Overview
Forever-Store is a distributed file storage system written in Go. The project is currently a work in progress and focuses on building a peer-to-peer (P2P) architecture for file storage with future support for encryption, REST/CLI interfaces, and containerized deployment.

## Current Implementation

### P2P Package
- Implemented a dedicated `p2p` package for peer communication and data transportation.
- Defined and implemented `Peer` and `Transporter` interfaces to abstract network-level operations.
- Handles TCP peer connections, including:
  - Writing to a connection.
  - Consuming data from a connection.
  - General network transport logic.

### Storage
- Added logic for file storage using Content-Addressable Storage (CAS).
- Implemented functionality for converting filenames into an encrypted form (initial work).
- Establishes the foundation for secure, addressable file management.

## Future Work
- **Encryption**: Implement complete encryption for stored data files.
- **Integration**: Connect P2P networking, CAS storage, and encryption into a unified workflow.
- **Client Interaction**: Develop a client interface, either as:
  - A REST API for external integrations.
  - A CLI tool for direct interaction.
- **Deployment**: Containerize the system using Docker for easier distribution and testing.

## Status
The system is under active development. Current focus is on extending encryption and integrating core components.
